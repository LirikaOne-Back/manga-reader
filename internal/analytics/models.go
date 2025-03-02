package analytics

// TopMangaEntry представляет элемент рейтинга манги
type TopMangaEntry struct {
	MangaID int64 `json:"manga_id"`
	Views   int64 `json:"views"`
}

// MangaWithViews представляет мангу с информацией о просмотрах
type MangaWithViews struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Views       int64  `json:"views"`
}

// ChapterWithViews представляет главу с информацией о просмотрах
type ChapterWithViews struct {
	ID      int64  `json:"id"`
	MangaID int64  `json:"manga_id"`
	Number  int    `json:"number"`
	Title   string `json:"title"`
	Views   int64  `json:"views"`
}
