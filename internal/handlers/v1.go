package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"kphotos/internal/cache"
	"kphotos/internal/db"
	"kphotos/internal/models"
	"kphotos/internal/utils"

	"github.com/go-chi/chi/v5"

	_ "golang.org/x/image/webp"
)

func UploadPhoto(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("photo")
	if err != nil {
		http.Error(w, "Gagal membaca file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// pastikan folder ada
	os.MkdirAll("photos/originals", 0755)
	os.MkdirAll("photos/thumbs", 0755)

	// simpan original
	dst := filepath.Join("photos/originals", header.Filename)
	out, err := os.Create(dst)
	if err != nil {
		http.Error(w, "Gagal menyimpan file", http.StatusInternalServerError)
		return
	}
	if _, err := io.Copy(out, file); err != nil {
		http.Error(w, "Gagal menulis file", http.StatusInternalServerError)
		return
	}
	out.Close()

	// deteksi ekstensi
	ext := strings.ToLower(filepath.Ext(header.Filename))
	srcPath := dst

	// 🔄 jika file HEIC/HEIF, konversi ke JPG dulu agar bisa diproses imaging
	if ext == ".heic" || ext == ".heif" {
		tempPath := strings.TrimSuffix(dst, ext) + "_converted.jpg"
		cmd := exec.Command("heif-convert", dst, tempPath)
		if err := cmd.Run(); err != nil {
			http.Error(w, fmt.Sprintf("Gagal konversi HEIC: %v", err), http.StatusInternalServerError)
			return
		}
		srcPath = tempPath
	}

	// buat thumbnail
	thumbPath := filepath.Join("photos/thumbs", header.Filename+".jpg")
	if err := utils.GenerateThumbnail(srcPath, thumbPath); err != nil {
		http.Error(w, fmt.Sprintf("Gagal membuat thumbnail: %v", err), http.StatusInternalServerError)
		return
	}

	// hapus file sementara (hasil konversi)
	if srcPath != dst {
		os.Remove(srcPath)
	}

	// simpan metadata ke DB
	res, err := db.DB.Exec(`
		INSERT INTO photos (file_name, file_path, thumb_path, size)
		VALUES (?, ?, ?, ?)`,
		header.Filename, dst, thumbPath, header.Size,
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("DB error: %v", err), http.StatusInternalServerError)
		return
	}

	id, _ := res.LastInsertId()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":       id,
		"fileName": header.Filename,
	})
}

func ListPhotos(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query(`SELECT id, file_name, file_path, thumb_path, created_at FROM photos ORDER BY created_at DESC`)
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var photos []models.Photo
	for rows.Next() {
		var p models.Photo
		rows.Scan(&p.ID, &p.FileName, &p.FilePath, &p.ThumbPath, &p.CreatedAt)
		photos = append(photos, p)
	}

	json.NewEncoder(w).Encode(photos)
}
func ServeOriginal(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var filePath string
	err := db.DB.QueryRow("SELECT file_path FROM photos WHERE id = ?", id).Scan(&filePath)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Foto tidak ditemukan", http.StatusNotFound)
		} else {
			http.Error(w, "Kesalahan database", http.StatusInternalServerError)
		}
		return
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	if ext != ".heic" && ext != ".heif" {
		http.ServeFile(w, r, filePath)
		return
	}

	// Cek cache
	cachedPath := cache.GetCachedPath(filePath, ".jpg")
	if utils.FileExists(cachedPath) {
		http.ServeFile(w, r, cachedPath)
		return
	}

	// Buat direktori cache jika belum ada
	os.MkdirAll(filepath.Dir(cachedPath), 0755)

	// --- Convert HEIC ke JPEG sementara ---
	tmpJPG := strings.TrimSuffix(cachedPath, ".jpg") + "_conv.jpg"
	cmd := exec.Command("heif-convert", filePath, tmpJPG)
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Println("heif-convert error:", string(out))
		http.Error(w, "Gagal mengonversi HEIC", http.StatusInternalServerError)
		return
	}

	// Pastikan file hasil konversi ada
	if !utils.FileExists(tmpJPG) {
		http.Error(w, "File hasil konversi tidak ditemukan", http.StatusInternalServerError)
		return
	}

	// --- Perbaiki orientasi hasil konversi ---
	imgFile, err := utils.OpenFile(tmpJPG)
	if err != nil {
		http.Error(w, "Gagal membaca hasil konversi", http.StatusInternalServerError)
		return
	}
	defer imgFile.Close()

	img, _, err := image.Decode(imgFile)
	if err != nil {
		http.Error(w, "Gagal decode hasil konversi", http.StatusInternalServerError)
		return
	}

	orientation := utils.GetOrientation(filePath) // baca orientasi dari file HEIC
	img = utils.FixOrientation(img, orientation)

	// Encode ulang ke buffer
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		http.Error(w, "Gagal encode ke JPEG", http.StatusInternalServerError)
		return
	}

	// Simpan ke cache (overwrite)
	if err := cache.Save(cachedPath, buf.Bytes()); err != nil {
		fmt.Println("Gagal simpan cache:", err)
	}

	// Hapus file konversi sementara
	os.Remove(tmpJPG)

	// Kirim ke browser
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.WriteHeader(http.StatusOK)
	w.Write(buf.Bytes())
}

func ServeThumb(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var path string
	db.DB.QueryRow("SELECT thumb_path FROM photos WHERE id=?", id).Scan(&path)
	http.ServeFile(w, r, path)
}

// GET /albums
func listAlbums(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query("SELECT * FROM albums ORDER BY created_at DESC")
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var albums []models.Album
	for rows.Next() {
		var a models.Album
		if err := rows.Scan(&a.ID, &a.Name, &a.Description, &a.Cover, &a.CreatedAt); err != nil {
			continue
		}
		albums = append(albums, a)
	}

	jsonResponse(w, albums)
}

// POST /albums
func createAlbum(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.Name == "" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	result, err := db.DB.Exec("INSERT INTO albums (name, description) VALUES (?,?)", payload.Name, payload.Description)
	if err != nil {
		errResponse := fmt.Sprintf("Failed to create album: %s", err.Error())
		http.Error(w, errResponse, http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	jsonResponse(w, map[string]any{"id": id, "name": payload.Name})
}

// GET /albums/{id}/photos
func getAlbumPhotos(w http.ResponseWriter, r *http.Request) {
	albumID := chi.URLParam(r, "id")

	query := `
		SELECT p.id, p.name, p.file_path
		FROM photos p
		JOIN album_photos ap ON p.id = ap.photo_id
		WHERE ap.album_id = ?;
	`

	rows, err := db.DB.Query(query, albumID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type Photo struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		FilePath string `json:"file_path"`
	}

	var photos []Photo
	for rows.Next() {
		var p Photo
		if err := rows.Scan(&p.ID, &p.Name, &p.FilePath); err != nil {
			continue
		}
		photos = append(photos, p)
	}

	jsonResponse(w, photos)
}

// POST /albums/{id}/add/{photoId}
func addPhotoToAlbum(w http.ResponseWriter, r *http.Request) {
	albumID := chi.URLParam(r, "id")
	photoID := chi.URLParam(r, "photoId")

	_, err := db.DB.Exec("INSERT OR IGNORE INTO album_photos (album_id, photo_id) VALUES (?, ?)", albumID, photoID)
	if err != nil {
		http.Error(w, "Failed to add photo", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, map[string]string{"status": "added"})
}

// DELETE /albums/{id}/remove/{photoId}
func removePhotoFromAlbum(w http.ResponseWriter, r *http.Request) {
	albumID := chi.URLParam(r, "id")
	photoID := chi.URLParam(r, "photoId")

	_, err := db.DB.Exec("DELETE FROM album_photos WHERE album_id = ? AND photo_id = ?", albumID, photoID)
	if err != nil {
		http.Error(w, "Failed to remove photo", http.StatusInternalServerError)
		return
	}

	jsonResponse(w, map[string]string{"status": "removed"})
}

// DELETE /albums/{id}
func deleteAlbum(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	tx, err := db.DB.Begin()
	if err != nil {
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}

	// hapus relasi dulu
	if _, err := tx.Exec("DELETE FROM album_photos WHERE album_id = ?", id); err != nil {
		tx.Rollback()
		http.Error(w, "Failed to delete relations", http.StatusInternalServerError)
		return
	}

	// hapus album
	if _, err := tx.Exec("DELETE FROM albums WHERE id = ?", id); err != nil {
		tx.Rollback()
		http.Error(w, "Failed to delete album", http.StatusInternalServerError)
		return
	}

	tx.Commit()
	jsonResponse(w, map[string]string{"status": "deleted"})
}

// Helper JSON
func jsonResponse(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")

	// Jika data adalah slice kosong, return [] bukan null
	switch v := data.(type) {
	case nil:
		w.Write([]byte("[]"))
		return
	case []any:
		if len(v) == 0 {
			w.Write([]byte("[]"))
			return
		}
	case []models.Album:
		if len(v) == 0 {
			w.Write([]byte("[]"))
			return
		}
	case []models.Photo:
		if len(v) == 0 {
			w.Write([]byte("[]"))
			return
		}
	}

	json.NewEncoder(w).Encode(data)
}
