package sqlite

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log/slog"
	"manga-reader/internal/db"
	"manga-reader/models"
)

type SQLiteChapterRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewChapterRepository(conn *sql.DB, logger *slog.Logger) db.ChapterRepository {
	repo := &SQLiteChapterRepository{db: conn, logger: logger}
	if err := repo.initSchema(); err != nil {
		logger.Error("Ошибка создания схемы для глав", "err", err)
	}
	return repo
}

func (r *SQLiteChapterRepository) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS chapter (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		manga_id INTEGER NOT NULL,
		number INTEGER NOT NULL,
		title TEXT NOT NULL,
		FOREIGN KEY(manga_id) REFERENCES manga(id)
	);`
	_, err := r.db.Exec(schema)
	if err != nil {
		r.logger.Error("Ошибка создания таблицы chapter", "err", err)
	}
	return err
}

func (r *SQLiteChapterRepository) Create(ch *models.Chapter) (int64, error) {
	result, err := r.db.Exec("INSERT INTO chapter (manga_id, number, title) VALUES (?, ?, ?)", ch.MangaID, ch.Number, ch.Title)
	if err != nil {
		r.logger.Error("Ошибка вставки главы", "err", err)
		return 0, err
	}
	return result.LastInsertId()
}

func (r *SQLiteChapterRepository) GetByID(id int64) (*models.Chapter, error) {
	row := r.db.QueryRow("SELECT id, manga_id, number, title FROM chapter WHERE id = ?", id)
	ch := &models.Chapter{}
	if err := row.Scan(&ch.ID, &ch.MangaID, &ch.Number, &ch.Title); err != nil {
		r.logger.Error("Ошибка получения главы", "err", err)
		return nil, err
	}
	return ch, nil
}

func (r *SQLiteChapterRepository) ListByManga(mangaID int64) ([]*models.Chapter, error) {
	rows, err := r.db.Query("SELECT id, manga_id, number, title FROM chapter WHERE manga_id = ? ORDER BY number", mangaID)
	if err != nil {
		r.logger.Error("Ошибка получения списка глав", "err", err)
		return nil, err
	}
	defer rows.Close()

	var chapters []*models.Chapter
	for rows.Next() {
		ch := &models.Chapter{}
		if err := rows.Scan(&ch.ID, &ch.MangaID, &ch.Number, &ch.Title); err != nil {
			r.logger.Error("Ошибка сканирования главы", "err", err)
			return nil, err
		}
		chapters = append(chapters, ch)
	}
	return chapters, nil
}

func (r *SQLiteChapterRepository) Update(ch *models.Chapter) error {
	result, err := r.db.Exec("UPDATE chapter SET number = ?, title = ? WHERE id = ?", ch.Number, ch.Title, ch.ID)
	if err != nil {
		r.logger.Error("Ошибка обновления главы", "err", err)
		return err
	}
	if affected, err := result.RowsAffected(); err != nil || affected == 0 {
		err = fmt.Errorf("глава с id %d не найдена", ch.ID)
		r.logger.Error("Ошибка обновления главы", "err", err)
		return err
	}
	return nil
}

func (r *SQLiteChapterRepository) Delete(id int64) error {
	result, err := r.db.Exec("DELETE FROM chapter WHERE id = ?", id)
	if err != nil {
		r.logger.Error("Ошибка удаления главы", "err", err)
		return err
	}
	if affected, err := result.RowsAffected(); err != nil || affected == 0 {
		err = fmt.Errorf("глава с id %d не найдена", id)
		r.logger.Error("Ошибка удаления главы", "err", err)
		return err
	}
	return nil
}
