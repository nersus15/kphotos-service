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

	// Albums
	r.Get("/albums", listAlbums)
	r.Post("/albums", createAlbum)
	r.Get("/albums/{id}/photos", getAlbumPhotos)
	r.Post("/albums/{id}/add/{photoId}", addPhotoToAlbum)
	r.Delete("/albums/{id}/remove/{photoId}", removePhotoFromAlbum)
	r.Delete("/albums/{id}", deleteAlbum)
	return r
}
