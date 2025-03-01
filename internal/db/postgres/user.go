package postgres

import (
	"database/sql"
	"log/slog"
	"manga-reader/models"
)

type PostgresUserRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewUserRepository(db *sql.DB, logger *slog.Logger) *PostgresUserRepository {
	return &PostgresUserRepository{
		db:     db,
		logger: logger,
	}
}

func (r *PostgresUserRepository) Create(user *models.User) (int64, error) {
	var id int64
	err := r.db.QueryRow(
		"INSERT INTO users (username, password) VALUES ($1, $2) RETURNING id",
		user.Username, user.Password,
	).Scan(&id)

	if err != nil {
		r.logger.Error("Ошибка создания пользователя в PostgreSQL", "err", err)
		return 0, err
	}

	return id, nil
}

func (r *PostgresUserRepository) GetByUsername(username string) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRow(
		"SELECT id, username, password FROM users WHERE username = $1",
		username,
	).Scan(&user.ID, &user.Username, &user.Password)

	if err != nil {
		r.logger.Error("Ошибка получения пользователя по username из PostgreSQL", "err", err, "username", username)
		return nil, err
	}

	return user, nil
}
