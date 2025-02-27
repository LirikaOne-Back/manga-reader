package sqlite

import (
	"database/sql"
	"fmt"
	"log/slog"
	"manga-reader/models"
)

type SQLitePageRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewPageRepository(db *sql.DB, logger *slog.Logger) *SQLitePageRepository {
	repo := &SQLitePageRepository{
		db:     db,
		logger: logger,
	}
	if err := repo.initSchema(); err != nil {
		logger.Error("Ошибка создания схемы для страниц", "err", err)
	}
	return repo
}

func (r *SQLitePageRepository) initSchema() error {
	schema := `CREATE TABLE IF NOT EXISTS pages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    chapter_id INTEGER NOT NULL,
    number INTEGER NOT NULL,
    image_path TEXT NOT NULL,
    FOREIGN KEY(chapter_id) REFERENCES chapters(id));`
	_, err := r.db.Exec(schema)
	if err != nil {
		r.logger.Error("Ошибка создания таблицы pages", "err", err)
	}
	return err
}

func (r *SQLitePageRepository) Create(p *models.Page) (int64, error) {
	res, err := r.db.Exec("INSERT INTO pages (chapter_id, number, image_path) VALUES (?, ?, ?)", p.ChapterID, p.Number, p.ImagePath)
	if err != nil {
		r.logger.Error("Ошибка создания новой страницы", "err", err)
		return 0, err
	}
	return res.LastInsertId()
}

func (r *SQLitePageRepository) GetByID(id int64) (*models.Page, error) {
	row := r.db.QueryRow("SELECT id, chapter_id, number, image_path FROM pages WHERE id = ?", id)
	page := &models.Page{}
	if err := row.Scan(&page.ID, &page.ChapterID, &page.Number, &page.ImagePath); err != nil {
		r.logger.Error("Ошибка получения страницы", "err", err)
		return nil, err
	}
	return page, nil
}

func (r *SQLitePageRepository) ListByChapter(chapterID int64) ([]*models.Page, error) {
	rows, err := r.db.Query("SELECT id, chapter_id, number, image_path FROM pages WHERE chapter_id = ?", chapterID)
	if err != nil {
		r.logger.Error("Ошибка получения списка стрпаниц", "err", err)
		return nil, err
	}
	defer rows.Close()
	pages := []*models.Page{}
	for rows.Next() {
		page := &models.Page{}
		if err = rows.Scan(&page.ID, &page.ChapterID, &page.Number, &page.ImagePath); err != nil {
			r.logger.Error("Ошибка сканирования страницы", "err", err)
			return nil, err
		}
		pages = append(pages, page)
	}
	return pages, nil
}

func (r *SQLitePageRepository) Update(p *models.Page) error {
	res, err := r.db.Exec("UPDATE pages SET chapter_id = ?, number = ?, image_path = ? WHERE id = ?", p.ChapterID, p.Number, p.ImagePath, p.ID)
	if err != nil {
		r.logger.Error("Ошибка обновления страницы", "err", err)
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		r.logger.Error("Ошибка получения количенства измененных строк", "err", err)
		return err
	}
	if affected == 0 {
		err = fmt.Errorf("страница с id %d не найдена", p.ID)
		r.logger.Error("Ошибка обновления страницы", "err", err)
		return err
	}
	return nil
}

func (r *SQLitePageRepository) Delete(id int64) error {
	res, err := r.db.Exec("DELETE FROM pages  WHERE id = ?", id)
	if err != nil {
		r.logger.Error("Ошибка удаления страницы", "err", err)
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		r.logger.Error("Ошибка получения количенства удаленных строк", "err", err)
		return err
	}
	if affected == 0 {
		err = fmt.Errorf("страница с id %d не найдена", id)
		r.logger.Error("Ошибка удаления страницы", "err", err)
		return err
	}
	return nil
}
