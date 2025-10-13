package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"kphotos/internal/db"
	"kphotos/internal/models"
	"kphotos/internal/utils"

	"github.com/go-chi/chi/v5"
)

func UploadPhoto(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("photo")
	if err != nil {
		http.Error(w, "Gagal membaca file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	dst := filepath.Join("photos/originals", header.Filename)
	out, err := os.Create(dst)
	if err != nil {
		http.Error(w, "Gagal menyimpan file", http.StatusInternalServerError)
		return
	}
	io.Copy(out, file)
	out.Close()

	thumbPath := filepath.Join("photos/thumbs", header.Filename)
	if err := utils.GenerateThumbnail(dst, thumbPath); err != nil {
		http.Error(w, "Gagal membuat thumbnail", http.StatusInternalServerError)
		return
	}

	res, err := db.DB.Exec(`
		INSERT INTO photos (file_name, file_path, thumb_path, size)
		VALUES (?, ?, ?, ?)`,
		header.Filename, dst, thumbPath, header.Size)
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}

	id, _ := res.LastInsertId()
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
