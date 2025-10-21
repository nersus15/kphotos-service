package models

import (
	"database/sql"
)

type Album struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description sql.NullString `json:"-"`
	Cover       sql.NullString `json:"-"`
	CreatedAt   string         `json:"created_at"`
}
type AlbumResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Cover       string `json:"cover"`
	CreatedAt   string `json:"created_at"`
}
