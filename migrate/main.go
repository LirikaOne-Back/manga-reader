package main

import (
	"flag"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"log"
	"manga-reader/config"
	"os"
)

func main() {
	up := flag.Bool("up", false, "Применить миграции вперед")
	down := flag.Bool("down", false, "Откатить миграции назад")
	version := flag.Int("version", 0, "Перейти к конкретной версии миграции")

	flag.Parse()

	if !*up && !*down && *version == 0 {
		flag.Usage()
		os.Exit(1)
	}

	cfg := config.LoadConfig()
	if cfg.DBType != "postgres" {
		log.Fatal("Миграции поддерживаются только для PostgreSQL")
	}

	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.PgUser, cfg.PgPassword, cfg.PgHost, cfg.PgPort, cfg.PgDBName, cfg.PgSSLMode)

	m, err := migrate.New("file://migrations/postgres", connString)
	if err != nil {
		log.Fatalf("Ошибка создания миграции: %v", err)
	}
	if *up {
		if err := m.Up(); err != nil {
			log.Fatalf("Ошибка применения миграций: %v", err)
		}
		log.Println("Миграции применены успешно")
	} else if *down {
		if err := m.Down(); err != nil {
			log.Fatalf("Ошибка отката миграций: %v", err)
		}
		log.Println("Миграции откачены успешно")
	} else if *version > 0 {
		if err := m.Migrate(uint(*version)); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Ошибка перехода к версии %d: %v", *version, err)
		}
		log.Printf("Успешный переход к версии %d\n", *version)
	}
}
