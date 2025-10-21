package models

type Album struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Cover       string `json:"cover"`
	CreatedAt   string `json:"created_at"`
}
