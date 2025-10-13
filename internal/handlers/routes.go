package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func V1Routes() http.Handler {
	r := chi.NewRouter()

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("health check..."))
	})

	r.Post("/upload", UploadPhoto)
	r.Get("/photos", ListPhotos)
	r.Get("/media/{id}", ServeOriginal)
	r.Get("/thumb/{id}", ServeThumb)
	return r
}
