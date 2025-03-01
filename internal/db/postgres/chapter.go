package postgres

import (
	"database/sql"
	"fmt"
	"log/slog"

	"manga-reader/internal/db"
	"manga-reader/models"
)

type PostgresChapterRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewChapterRepository(db *sql.DB, logger *slog.Logger) db.ChapterRepository {
	return &PostgresChapterRepository{db: db, logger: logger}
}

func (r *PostgresChapterRepository) Create(ch *models.Chapter) (int64, error) {
	var id int64
	err := r.db.QueryRow(
		"INSERT INTO chapters (manga_id, number, title) VALUES ($1, $2, $3) RETURNING id",
		ch.MangaID, ch.Number, ch.Title,
	).Scan(&id)

	if err != nil {
		r.logger.Error("Ошибка вставки главы в PostgreSQL", "err", err)
		return 0, err
	}

	return id, nil
}

func (r *PostgresChapterRepository) GetByID(id int64) (*models.Chapter, error) {
	ch := &models.Chapter{}
	err := r.db.QueryRow(
		"SELECT id, manga_id, number, title FROM chapters WHERE id = $1",
		id,
	).Scan(&ch.ID, &ch.MangaID, &ch.Number, &ch.Title)

	if err != nil {
		r.logger.Error("Ошибка получения главы из PostgreSQL", "err", err, "id", id)
		return nil, err
	}

	return ch, nil
}

func (r *PostgresChapterRepository) ListByManga(mangaID int64) ([]*models.Chapter, error) {
	rows, err := r.db.Query(
		"SELECT id, manga_id, number, title FROM chapters WHERE manga_id = $1 ORDER BY number",
		mangaID,
	)

	if err != nil {
		r.logger.Error("Ошибка получения списка глав из PostgreSQL", "err", err, "manga_id", mangaID)
		return nil, err
	}
	defer rows.Close()

	var chapters []*models.Chapter
	for rows.Next() {
		ch := &models.Chapter{}
		if err := rows.Scan(&ch.ID, &ch.MangaID, &ch.Number, &ch.Title); err != nil {
			r.logger.Error("Ошибка сканирования главы из PostgreSQL", "err", err)
			return nil, err
		}
		chapters = append(chapters, ch)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("Ошибка итерации по результатам из PostgreSQL", "err", err)
		return nil, err
	}

	return chapters, nil
}

func (r *PostgresChapterRepository) Update(ch *models.Chapter) error {
	result, err := r.db.Exec(
		"UPDATE chapters SET number = $1, title = $2 WHERE id = $3",
		ch.Number, ch.Title, ch.ID,
	)

	if err != nil {
		r.logger.Error("Ошибка обновления главы в PostgreSQL", "err", err, "id", ch.ID)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Ошибка получения количества обновленных строк в PostgreSQL", "err", err)
		return err
	}

	if rowsAffected == 0 {
		err = fmt.Errorf("глава с id %d не найдена", ch.ID)
		r.logger.Error("Глава не найдена для обновления в PostgreSQL", "id", ch.ID)
		return err
	}

	return nil
}

func (r *PostgresChapterRepository) Delete(id int64) error {
	result, err := r.db.Exec("DELETE FROM chapters WHERE id = $1", id)
	if err != nil {
		r.logger.Error("Ошибка удаления главы из PostgreSQL", "err", err, "id", id)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Ошибка получения количества удаленных строк в PostgreSQL", "err", err)
		return err
	}

	if rowsAffected == 0 {
		err = fmt.Errorf("глава с id %d не найдена", id)
		r.logger.Error("Глава не найдена для удаления в PostgreSQL", "id", id)
		return err
	}

	return nil
}
