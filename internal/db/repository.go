package db

import "manga-reader/models"

// MangaRepository описывает операции над мангой.
type MangaRepository interface {
	Create(m *models.Manga) (int64, error)
	GetByID(id int64) (*models.Manga, error)
	List() ([]*models.Manga, error)
	Update(m *models.Manga) error
	Delete(id int64) error
}

// ChapterRepository описывает операции над главами манги.
type ChapterRepository interface {
	Create(ch *models.Chapter) (int64, error)
	GetByID(id int64) (*models.Chapter, error)
	ListByManga(mangaID int64) ([]*models.Chapter, error)
	Update(ch *models.Chapter) error
	Delete(id int64) error
}

// PageRepository описывает операции над страницами глав.
type PageRepository interface {
	Create(p *models.Page) (int64, error)
	GetByID(id int64) (*models.Page, error)
	ListByChapter(chapterID int64) ([]*models.Page, error)
	Update(p *models.Page) error
	Delete(id int64) error
}

// UserRepository описывает операции над пользователями.
type UserRepository interface {
	Create(user *models.User) (int64, error)
	GetByUsername(username string) (*models.User, error)
}
