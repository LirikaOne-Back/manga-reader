package models

type Page struct {
	ID        int64  `json:"id"`
	ChapterID int64  `json:"chapter_id"`
	Number    int    `json:"number"`
	ImagePath string `json:"image_path"`
}
