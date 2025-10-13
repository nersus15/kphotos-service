package models

type Photo struct {
	ID          int64  `json:"id"`
	FileName    string `json:"file_name"`
	FilePath    string `json:"file_path"`
	ThumbPath   string `json:"thumb_path"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	Size        int64  `json:"size"`
	ExifTakenAt string `json:"exif_taken_at"`
	CreatedAt   string `json:"created_at"`
}
