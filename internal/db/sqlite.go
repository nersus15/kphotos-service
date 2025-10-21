package db

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB(path string) {
	var err error
	DB, err = sql.Open("sqlite", path)
	if err != nil {
		log.Fatal(err)
	}

	createTable := `
	CREATE TABLE IF NOT EXISTS photos (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		file_name TEXT NOT NULL,
		file_path TEXT NOT NULL,
		thumb_path TEXT,
		width INTEGER,
		height INTEGER,
		size INTEGER,
		exif_taken_at DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err = DB.Exec(createTable)
	if err != nil {
		log.Fatal("Gagal membuat tabel:", err)
	}

	// Create table for albums
	createTable = `
		CREATE TABLE IF NOT EXISTS albums (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		description TEXT NULL DEFAULT NULL,
		cover TEXT NULL DEFAULT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS album_photos (
		album_id INTEGER,
		photo_id INTEGER,
		PRIMARY KEY (album_id, photo_id),
		FOREIGN KEY (album_id) REFERENCES albums(id),
		FOREIGN KEY (photo_id) REFERENCES photos(id)
		);
	`
	_, err = DB.Exec(createTable)
	if err != nil {
		log.Fatal("Gagal membuat tabel:", err)
	}

	DB.Exec("PRAGMA journal_mode=WAL;")
}
