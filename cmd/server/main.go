package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"manga-reader/config"
	"manga-reader/internal/cache"
	"manga-reader/internal/db"
	"manga-reader/internal/db/sqlite"
	"manga-reader/internal/handlers"
	"manga-reader/internal/logger"
	"manga-reader/internal/middleware"
)

func main() {
	cfg := config.LoadConfig()
	log := logger.NewLogger()

	var mangaRepo db.MangaRepository
	var err error
	switch cfg.DBType {
	case "sqlite":
		mangaRepo, err = sqlite.NewMangaRepository(cfg.DBSource, log)
	case "postgres":
		log.Error("Postgres пока не поддерживается")
		return
	default:
		log.Error("Неизвестный тип базы данных", "string", cfg.DBType)
		return
	}
	if err != nil {
		log.Error("Ошибка инициализации базы данных", "err", err)
		return
	}

	redisCache := cache.NewRedisCache(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB, log)

	mangaHandler := &handlers.MangaHandler{
		Repo:   mangaRepo,
		Logger: log,
		Cache:  redisCache,
	}
	var chapterRepo db.ChapterRepository
	if sqliteRepo, ok := mangaRepo.(*sqlite.SQLiteMangaRepository); ok {
		chapterRepo = sqlite.NewChapterRepository(sqliteRepo.GetDB(), log)
	}
	chapterHandler := &handlers.ChapterHandler{
		Repo:   chapterRepo,
		Logger: log,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", handlers.HealthHandler)

	handlers.RegisterMangaRoutes(mux, mangaHandler, chapterHandler)
	handlers.RegisterChapterRoutes(mux, chapterHandler)

	handler := middleware.RecoveryMiddleware(log, middleware.LoggingMiddleware(log, mux))

	server := &http.Server{
		Addr:         cfg.ServerAddress,
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	go func() {
		log.Info("Запуск сервера на " + cfg.ServerAddress)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("Ошибка сервера", "err", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Получен сигнал завершения, закрываем сервер...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err = server.Shutdown(ctx); err != nil {
		log.Error("Ошибка при завершении работы сервера", "err", err)
	}
	log.Info("Сервер завершил работу")
}
