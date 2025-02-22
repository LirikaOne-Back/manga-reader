package models

type Chapter struct {
	ID      int64  `json:"id"`
	MangaID int64  `json:"manga_id"`
	Number  int    `json:"number"`
	Title   string `json:"title"`
}
