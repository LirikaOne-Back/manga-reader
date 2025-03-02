package analytics

import (
	"context"
	"fmt"
	"log/slog"
	"manga-reader/internal/cache"
	"strconv"
	"time"
)

const (
	// Префиксы ключей Redis
	mangaViewsPrefix   = "views:manga:"
	chapterViewsPrefix = "views:chapter:"
	pageViewsPrefix    = "views:page:"

	// Ключи для рейтингов
	topMangaKey        = "ranking:manga"
	topMangaDailyKey   = "ranking:manga:daily"
	topMangaWeeklyKey  = "ranking:manga:weekly"
	topMangaMonthlyKey = "ranking:manga:monthly"

	// Время жизни ключей
	dailyExpire   = 24 * time.Hour
	weeklyExpire  = 7 * 24 * time.Hour
	monthlyExpire = 30 * 24 * time.Hour
)

type AnalyticsService struct {
	cache  cache.Cache
	logger *slog.Logger
}

func NewAnalyticsService(cache cache.Cache, logger *slog.Logger) *AnalyticsService {
	return &AnalyticsService{
		cache:  cache,
		logger: logger,
	}
}

func (s *AnalyticsService) RecordMangaView(ctx context.Context, mangaID int64) error {
	mangaKey := fmt.Sprintf("%s%d", mangaViewsPrefix, mangaID)

	_, err := s.cache.Incr(ctx, mangaKey)
	if err != nil {
		s.logger.Error("Ошибка инкремента счетчика просмотров манги", "manga_id", mangaID, "err", err)
		return err
	}

	_, err = s.cache.ZIncrBy(ctx, topMangaKey, 1, strconv.FormatInt(mangaID, 10))
	if err != nil {
		s.logger.Error("Ошибка обновления рейтинга манги", "manga_id", mangaID, "err", err)
		return err
	}

	_, err = s.cache.ZIncrBy(ctx, topMangaDailyKey, 1, strconv.FormatInt(mangaID, 10))
	if err != nil {
		s.logger.Error("Ошибка обновления дневного рейтинга манги", "manga_id", mangaID, "err", err)
		return err
	}

	_, err = s.cache.ZIncrBy(ctx, topMangaWeeklyKey, 1, strconv.FormatInt(mangaID, 10))
	if err != nil {
		s.logger.Error("Ошибка обновления недельного рейтинга манги", "manga_id", mangaID, "err", err)
		return err
	}

	_, err = s.cache.ZIncrBy(ctx, topMangaMonthlyKey, 1, strconv.FormatInt(mangaID, 10))
	if err != nil {
		s.logger.Error("Ошибка обновления месячного рейтинга манги", "manga_id", mangaID, "err", err)
		return err
	}
	return nil
}

func (s *AnalyticsService) RecordChapterView(ctx context.Context, chapterID, mangaID int64) error {
	mangaKey := fmt.Sprintf("%s%d", chapterViewsPrefix, chapterID)

	_, err := s.cache.Incr(ctx, mangaKey)
	if err != nil {
		s.logger.Error("Ошибка инкремента счетчика просмотров главы", "chapter_id", chapterID, "err", err)
		return err
	}

	return s.RecordMangaView(ctx, mangaID)
}

func (s *AnalyticsService) RecordPageView(ctx context.Context, pageID, chapterID, mangaID int64) error {
	mangaKey := fmt.Sprintf("%s%d", pageViewsPrefix, pageID)

	_, err := s.cache.Incr(ctx, mangaKey)
	if err != nil {
		s.logger.Error("Ошибка инкремента счетчика просмотров страницы", "page_id", pageID, "err", err)
		return err
	}

	return s.RecordChapterView(ctx, chapterID, mangaID)
}

func (s *AnalyticsService) GetMangaView(ctx context.Context, mangaID int64) (int64, error) {
	mangaKey := fmt.Sprintf("%s%d", mangaViewsPrefix, mangaID)

	viewsStr, err := s.cache.Get(ctx, mangaKey)
	if err != nil {
		if err.Error() == "redis: nil" {
			return 0, nil
		}
		s.logger.Error("Ошибка получения счетчика просмотров манги", "manga_id", mangaID, "err", err)
		return 0, err
	}

	views, err := strconv.ParseInt(viewsStr, 10, 64)
	if err != nil {
		s.logger.Error("Ошибка парсинга счетчика просмотров манги", "manga_id", mangaID, "err", err)
		return 0, err
	}

	return views, nil
}

func (s *AnalyticsService) GetTopManga(ctx context.Context, period string, limit int64) ([]TopMangaEntry, error) {
	var key string
	switch period {
	case "day":
		key = topMangaDailyKey
	case "week":
		key = topMangaWeeklyKey
	case "month":
		key = topMangaMonthlyKey
	default:
		key = topMangaKey
	}

	scoreMap, err := s.cache.ZRevRangeWithScores(ctx, key, 0, limit-1)
	if err != nil {
		s.logger.Error("Ошибка получения топ манги", "period", period, "err", err)
		return nil, err
	}

	var results []TopMangaEntry
	for member, score := range scoreMap {
		mangaID, err := strconv.ParseInt(member, 10, 64)
		if err != nil {
			s.logger.Error("Ошибка парсинга ID манги", "member", member, "err", err)
			continue
		}

		results = append(results, TopMangaEntry{
			MangaID: mangaID,
			Views:   int64(score),
		})
	}
	return results, nil
}

func (s *AnalyticsService) InitializeDailyStats(ctx context.Context) error {
	err := s.cache.Delete(ctx, topMangaDailyKey)
	if err != nil {
		s.logger.Error("Ошибка удаления дневного рейтинга", "err", err)
		return err
	}
	err = s.cache.Set(ctx, topMangaDailyKey+":expire", "1", dailyExpire)
	if err != nil {
		s.logger.Error("Ошибка установки времени жизни для дневного рейтинга", "err", err)
		return err
	}
	return nil
}

func (s *AnalyticsService) InitializeWeeklyStats(ctx context.Context) error {
	err := s.cache.Delete(ctx, topMangaWeeklyKey)
	if err != nil {
		s.logger.Error("Ошибка удаления недельного рейтинга", "err", err)
		return err
	}

	err = s.cache.Set(ctx, topMangaWeeklyKey+":expire", "1", weeklyExpire)
	if err != nil {
		s.logger.Error("Ошибка установки времени жизни для недельного рейтинга", "err", err)
		return err
	}

	return nil
}

func (s *AnalyticsService) InitializeMonthlyStats(ctx context.Context) error {
	err := s.cache.Delete(ctx, topMangaMonthlyKey)
	if err != nil {
		s.logger.Error("Ошибка удаления месячного рейтинга", "err", err)
		return err
	}

	err = s.cache.Set(ctx, topMangaMonthlyKey+":expire", "1", monthlyExpire)
	if err != nil {
		s.logger.Error("Ошибка установки времени жизни для месячного рейтинга", "err", err)
		return err
	}

	return nil
}
