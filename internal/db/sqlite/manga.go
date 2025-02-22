package sqlite

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log/slog"
	"manga-reader/internal/db"
	"manga-reader/models"
)

type SQLiteMangaRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewMangaRepository(dataSourceName string, logger *slog.Logger) (db.MangaRepository, error) {
	conn, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, err
	}
	if err = conn.Ping(); err != nil {
		return nil, err
	}

	repo := &SQLiteMangaRepository{db: conn, logger: logger}
	if err := repo.initSchema(); err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *SQLiteMangaRepository) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS manga (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		description TEXT
	);`
	_, err := r.db.Exec(schema)
	if err != nil {
		r.logger.Error("Ошибка создания схемы таблицы manga", "err", err)
	}
	return err
}

func (r *SQLiteMangaRepository) Create(m *models.Manga) (int64, error) {
	result, err := r.db.Exec("INSERT INTO manga (title, description) VALUES (?, ?)", m.Title, m.Description)
	if err != nil {
		r.logger.Error("Ошибка вставки манги", "err", err)
		return 0, err
	}
	return result.LastInsertId()
}

func (r *SQLiteMangaRepository) GetByID(id int64) (*models.Manga, error) {
	row := r.db.QueryRow("SELECT id, title, description FROM manga WHERE id = ?", id)
	m := &models.Manga{}
	if err := row.Scan(&m.ID, &m.Title, &m.Description); err != nil {
		r.logger.Error("Ошибка получения манги", "err", err)
		return nil, err
	}
	return m, nil
}

func (r *SQLiteMangaRepository) List() ([]*models.Manga, error) {
	rows, err := r.db.Query("SELECT id, title, description FROM manga")
	if err != nil {
		r.logger.Error("Ошибка получения списка манги", "err", err)
		return nil, err
	}
	defer rows.Close()

	var mangas []*models.Manga
	for rows.Next() {
		m := &models.Manga{}
		if err := rows.Scan(&m.ID, &m.Title, &m.Description); err != nil {
			r.logger.Error("Ошибка сканирования строки", "err", err)
			return nil, err
		}
		mangas = append(mangas, m)
	}
	return mangas, nil
}

func (r *SQLiteMangaRepository) Update(m *models.Manga) error {
	result, err := r.db.Exec("UPDATE manga SET title = ?, description = ? WHERE id = ?", m.Title, m.Description, m.ID)
	if err != nil {
		r.logger.Error("Ошибка обновления манги", "err", err)
		return err
	}
	if affected, err := result.RowsAffected(); err != nil || affected == 0 {
		r.logger.Error("Манга не найдена для обновления", "err", err)
		return err
	}
	return nil
}

func (r *SQLiteMangaRepository) Delete(id int64) error {
	result, err := r.db.Exec("DELETE FROM manga WHERE id = ?", id)
	if err != nil {
		r.logger.Error("Ошибка удаления манги", "err", err)
		return err
	}
	if affected, err := result.RowsAffected(); err != nil || affected == 0 {
		r.logger.Error("Манга не найдена для удаления", "err", err)
		return err
	}
	return nil
}

func (r *SQLiteMangaRepository) GetDB() *sql.DB {
	return r.db
}
