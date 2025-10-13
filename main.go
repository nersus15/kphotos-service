package main

import (
	"log"
	"net/http"
	"os"

	"kphotos/internal/db"
	"kphotos/internal/handlers"

	"github.com/go-chi/chi/v5"
)

func main() {
	// Buat folder jika belum ada
	os.MkdirAll("photos/originals", 0755)
	os.MkdirAll("photos/thumbs", 0755)
	os.MkdirAll("data", 0755)

	// Inisialisasi database SQLite
	db.InitDB("data/photos.db")

	r := chi.NewRouter()

	// Route API
	r.Mount("/api/v1", handlers.V1Routes())

	log.Println("🚀 Server berjalan di http://localhost:8080")
	http.ListenAndServe(":8080", r)
}
