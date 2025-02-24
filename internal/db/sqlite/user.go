package sqlite

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log/slog"
	"manga-reader/models"
)

type SQLiteUserRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewSQLiteUserRepository(db *sql.DB, logger *slog.Logger) *SQLiteUserRepository {
	repo := &SQLiteUserRepository{db: db, logger: logger}
	repo.initSchema()
	return repo
}

func (r *SQLiteUserRepository) initSchema() error {
	schema := `CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
	password TEXT NOT NULL);`
	_, err := r.db.Exec(schema)
	if err != nil {
		r.logger.Error("Ошибка создания таблицы users", "err", err)
	}
	return err
}

func (r *SQLiteUserRepository) Create(user *models.User) (int64, error) {
	result, err := r.db.Exec("INSERT INTO users (username, password) VALUES (?, ?)", user.Username, user.Password)
	if err != nil {
		r.logger.Error("Ошибка создания нового пользователя", "err", err)
	}
	return result.LastInsertId()
}

func (r *SQLiteUserRepository) GetByUsername(username string) (*models.User, error) {
	row := r.db.QueryRow("SELECT id, username, password FROM users WHERE username = ?", username)
	user := &models.User{}
	if err := row.Scan(&user.ID, &user.Username, &user.Password); err != nil {
		r.logger.Error("Ошибка получения пользователя по username", "err", err)
		return nil, err
	}
	return user, nil
}
