package models

import (
	"database/sql"
	"time"
)

type Album struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description sql.NullString `json:"-"`
	Cover       sql.NullString `json:"-"`
	CreatedAt   string         `json:"created_at"`
}
type AlbumResponse struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Cover       string    `json:"cover"`
	CreatedAt   time.Time `json:"created_at"`
}
