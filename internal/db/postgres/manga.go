package postgres

import (
	"database/sql"
	"log/slog"

	_ "github.com/lib/pq"
	"manga-reader/internal/db"
	"manga-reader/models"
)

type PostgresMangaRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewMangaRepository(connectionString string, logger *slog.Logger) (db.MangaRepository, error) {
	conn, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}
	if err = conn.Ping(); err != nil {
		return nil, err
	}

	repo := &PostgresMangaRepository{db: conn, logger: logger}
	return repo, nil
}

func (r *PostgresMangaRepository) Create(m *models.Manga) (int64, error) {
	var id int64
	err := r.db.QueryRow(
		"INSERT INTO manga (title, description) VALUES ($1, $2) RETURNING id",
		m.Title, m.Description,
	).Scan(&id)

	if err != nil {
		r.logger.Error("Ошибка вставки манги в PostgreSQL", "err", err)
		return 0, err
	}

	return id, nil
}

func (r *PostgresMangaRepository) GetByID(id int64) (*models.Manga, error) {
	m := &models.Manga{}
	err := r.db.QueryRow(
		"SELECT id, title, description FROM manga WHERE id = $1",
		id,
	).Scan(&m.ID, &m.Title, &m.Description)

	if err != nil {
		r.logger.Error("Ошибка получения манги из PostgreSQL", "err", err, "id", id)
		return nil, err
	}

	return m, nil
}

func (r *PostgresMangaRepository) List() ([]*models.Manga, error) {
	rows, err := r.db.Query("SELECT id, title, description FROM manga")
	if err != nil {
		r.logger.Error("Ошибка получения списка манги из PostgreSQL", "err", err)
		return nil, err
	}
	defer rows.Close()

	var mangas []*models.Manga
	for rows.Next() {
		m := &models.Manga{}
		if err := rows.Scan(&m.ID, &m.Title, &m.Description); err != nil {
			r.logger.Error("Ошибка сканирования строки из PostgreSQL", "err", err)
			return nil, err
		}
		mangas = append(mangas, m)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("Ошибка итерации по результатам из PostgreSQL", "err", err)
		return nil, err
	}

	return mangas, nil
}

func (r *PostgresMangaRepository) Update(m *models.Manga) error {
	result, err := r.db.Exec(
		"UPDATE manga SET title = $1, description = $2 WHERE id = $3",
		m.Title, m.Description, m.ID,
	)

	if err != nil {
		r.logger.Error("Ошибка обновления манги в PostgreSQL", "err", err, "id", m.ID)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Ошибка получения количества обновленных строк в PostgreSQL", "err", err)
		return err
	}

	if rowsAffected == 0 {
		r.logger.Error("Манга не найдена для обновления в PostgreSQL", "id", m.ID)
		return sql.ErrNoRows
	}

	return nil
}

func (r *PostgresMangaRepository) Delete(id int64) error {
	result, err := r.db.Exec("DELETE FROM manga WHERE id = $1", id)
	if err != nil {
		r.logger.Error("Ошибка удаления манги из PostgreSQL", "err", err, "id", id)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Ошибка получения количества удаленных строк в PostgreSQL", "err", err)
		return err
	}

	if rowsAffected == 0 {
		r.logger.Error("Манга не найдена для удаления в PostgreSQL", "id", id)
		return sql.ErrNoRows
	}

	return nil
}

func (r *PostgresMangaRepository) GetDB() *sql.DB {
	return r.db
}
