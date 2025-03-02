package main

import (
	"context"
	"errors"
	"manga-reader/internal/analytics"
	"manga-reader/internal/auth"
	"manga-reader/internal/db/postgres"
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

	auth.SetJWTSecret(cfg.JWTSecret)

	var mangaRepo db.MangaRepository
	var chapterRepo db.ChapterRepository
	var pageRepo db.PageRepository
	var userRepo db.UserRepository

	var err error
	switch cfg.DBType {
	case "sqlite":
		mangaRepo, err = sqlite.NewMangaRepository(cfg.DBSource, log)
		if err != nil {
			log.Error("Ошибка инициализации базы данных SQLite", "err", err)
			return
		}
		if sqliteRepo, ok := mangaRepo.(*sqlite.SQLiteMangaRepository); ok {
			chapterRepo = sqlite.NewChapterRepository(sqliteRepo.GetDB(), log)
			pageRepo = sqlite.NewPageRepository(sqliteRepo.GetDB(), log)
			userRepo = sqlite.NewSQLiteUserRepository(sqliteRepo.GetDB(), log)
		}
	case "postgres":
		connectionString := cfg.PostgresConnectionString()
		mangaRepo, err = postgres.NewMangaRepository(connectionString, log)
		if err != nil {
			log.Error("Ошибка инициализации базы данных PostgreSQL", "err", err)
			return
		}

		if pgRepo, ok := mangaRepo.(*postgres.PostgresMangaRepository); ok {
			chapterRepo = postgres.NewChapterRepository(pgRepo.GetDB(), log)
			pageRepo = postgres.NewPageRepository(pgRepo.GetDB(), log)
			userRepo = postgres.NewUserRepository(pgRepo.GetDB(), log)
		}
	default:
		log.Error("Неизвестный тип базы данных", "type", cfg.DBType)
		return
	}

	redisCache := cache.NewRedisCache(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB, log)

	analyticsService := analytics.NewAnalyticsService(redisCache, log)

	mangaHandler := &handlers.MangaHandler{
		Repo:      mangaRepo,
		Logger:    log,
		Cache:     redisCache,
		Analytics: analyticsService,
	}

	chapterHandler := &handlers.ChapterHandler{
		Repo:      chapterRepo,
		Logger:    log,
		Cache:     redisCache,
		Analytics: analyticsService,
	}

	pageHandler := &handlers.PageHandler{
		Repo:      pageRepo,
		Logger:    log,
		Cache:     redisCache,
		Analytics: analyticsService,
	}

	userHandler := &handlers.UserHandler{
		UserRepo: userRepo,
		Logger:   log,
	}

	analyticsHandler := &handlers.AnalyticsHandler{
		MangaRepo: mangaRepo,
		Analytics: analyticsService,
		Logger:    log,
	}

	mux := http.NewServeMux()
	mux.Handle("/", auth.AuthMiddleware(http.HandlerFunc(handlers.HealthHandler)))

	handlers.RegisterUserRoutes(mux, userHandler)
	handlers.RegisterMangaRoutes(mux, mangaHandler, chapterHandler)
	handlers.RegisterChapterRoutes(mux, chapterHandler)
	handlers.RegisterPageRoutes(mux, pageHandler)
	handlers.RegisterAnalyticsRoutes(mux, analyticsHandler)

	handler := middleware.RecoveryMiddleware(log, middleware.LoggingMiddleware(log, mux))

	server := &http.Server{
		Addr:         cfg.ServerAddress,
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
	ctx := context.Background()
	if err = analyticsService.InitializeDailyStats(ctx); err != nil {
		log.Error("Ошибка инициализации дневной статистики", "err", err)
	}
	if err = analyticsService.InitializeWeeklyStats(ctx); err != nil {
		log.Error("Ошибка инициализации недельной статистики", "err", err)
	}
	if err = analyticsService.InitializeMonthlyStats(ctx); err != nil {
		log.Error("Ошибка инициализации месячной статистики", "err", err)
	}

	go func() {
		log.Info("Запуск сервера на " + cfg.ServerAddress)
		if err = server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
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
