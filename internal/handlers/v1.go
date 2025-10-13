package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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
	var path string
	db.DB.QueryRow("SELECT file_path FROM photos WHERE id=?", id).Scan(&path)
	http.ServeFile(w, r, path)
}

func ServeThumb(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var path string
	db.DB.QueryRow("SELECT thumb_path FROM photos WHERE id=?", id).Scan(&path)
	http.ServeFile(w, r, path)
}
