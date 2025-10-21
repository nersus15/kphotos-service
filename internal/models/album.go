package models

type Album struct {
	ID          string `json:"id"`
	name        string `json:"name"`
	description string `json:"description"`
	CreatedAt   string `json:"created_at"`
}
