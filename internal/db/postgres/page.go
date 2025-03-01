package postgres

import (
	"database/sql"
	"fmt"
	"log/slog"
	"manga-reader/models"
)

type PostgresPageRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewPageRepository(db *sql.DB, logger *slog.Logger) *PostgresPageRepository {
	return &PostgresPageRepository{
		db:     db,
		logger: logger,
	}
}

func (r *PostgresPageRepository) Create(p *models.Page) (int64, error) {
	var id int64
	err := r.db.QueryRow(
		"INSERT INTO pages (chapter_id, number, image_path) VALUES ($1, $2, $3) RETURNING id",
		p.ChapterID, p.Number, p.ImagePath,
	).Scan(&id)

	if err != nil {
		r.logger.Error("Ошибка вставки страницы в PostgreSQL", "err", err)
		return 0, err
	}

	return id, nil
}

func (r *PostgresPageRepository) GetByID(id int64) (*models.Page, error) {
	page := &models.Page{}
	err := r.db.QueryRow(
		"SELECT id, chapter_id, number, image_path FROM pages WHERE id = $1",
		id,
	).Scan(&page.ID, &page.ChapterID, &page.Number, &page.ImagePath)

	if err != nil {
		r.logger.Error("Ошибка получения страницы из PostgreSQL", "err", err, "id", id)
		return nil, err
	}

	return page, nil
}

func (r *PostgresPageRepository) ListByChapter(chapterID int64) ([]*models.Page, error) {
	rows, err := r.db.Query(
		"SELECT id, chapter_id, number, image_path FROM pages WHERE chapter_id = $1 ORDER BY number",
		chapterID,
	)

	if err != nil {
		r.logger.Error("Ошибка получения списка страниц из PostgreSQL", "err", err, "chapter_id", chapterID)
		return nil, err
	}
	defer rows.Close()

	var pages []*models.Page
	for rows.Next() {
		page := &models.Page{}
		if err := rows.Scan(&page.ID, &page.ChapterID, &page.Number, &page.ImagePath); err != nil {
			r.logger.Error("Ошибка сканирования страницы из PostgreSQL", "err", err)
			return nil, err
		}
		pages = append(pages, page)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("Ошибка итерации по результатам из PostgreSQL", "err", err)
		return nil, err
	}

	return pages, nil
}

func (r *PostgresPageRepository) Update(p *models.Page) error {
	result, err := r.db.Exec(
		"UPDATE pages SET chapter_id = $1, number = $2, image_path = $3 WHERE id = $4",
		p.ChapterID, p.Number, p.ImagePath, p.ID,
	)

	if err != nil {
		r.logger.Error("Ошибка обновления страницы в PostgreSQL", "err", err, "id", p.ID)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Ошибка получения количества обновленных строк в PostgreSQL", "err", err)
		return err
	}

	if rowsAffected == 0 {
		err = fmt.Errorf("страница с id %d не найдена", p.ID)
		r.logger.Error("Страница не найдена для обновления в PostgreSQL", "id", p.ID)
		return err
	}

	return nil
}

func (r *PostgresPageRepository) Delete(id int64) error {
	result, err := r.db.Exec("DELETE FROM pages WHERE id = $1", id)
	if err != nil {
		r.logger.Error("Ошибка удаления страницы из PostgreSQL", "err", err, "id", id)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Ошибка получения количества удаленных строк в PostgreSQL", "err", err)
		return err
	}

	if rowsAffected == 0 {
		err = fmt.Errorf("страница с id %d не найдена", id)
		r.logger.Error("Страница не найдена для удаления в PostgreSQL", "id", id)
		return err
	}

	return nil
}
