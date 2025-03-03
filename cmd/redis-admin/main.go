package redis_admin

import (
	"context"
	"flag"
	"fmt"
	"log"
	"manga-reader/config"
	"manga-reader/internal/cache"
	"manga-reader/internal/logger"
	"os"
)

func main() {
	listCmd := flag.NewFlagSet("list", flag.ExitOnError)
	listPattern := listCmd.String("pattern", "*", "Шаблон ключей для вывода")

	getCmd := flag.NewFlagSet("get", flag.ExitOnError)
	getKey := getCmd.String("key", "", "Ключ для получения")

	deleteCmd := flag.NewFlagSet("delete", flag.ExitOnError)
	deleteKey := deleteCmd.String("key", "", "Ключ для удаления")

	flushCmd := flag.NewFlagSet("flush", flag.ExitOnError)
	flushConfirm := flushCmd.Bool("confirm", false, "Подтверждение операции очистки")

	zrangeCmd := flag.NewFlagSet("zrange", flag.ExitOnError)
	zrangeKey := zrangeCmd.String("key", "", "Ключ отсортированного множества")
	zrangeStart := zrangeCmd.Int64("start", 0, "Начальный индекс")
	zrangeStop := zrangeCmd.Int64("stop", -1, "Конечный индекс")

	if len(os.Args) < 2 {
		fmt.Println("Ожидается команда: list, get, delete, flush, zrange")
		os.Exit(1)
	}

	cfg := config.LoadConfig()
	log := logger.NewLogger()

	redisCache := cache.NewRedisCache(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB, log)
	ctx := context.Background()

	switch os.Args[1] {
	case "list":
		listCmd.Parse(os.Args[2:])
		listKeys(ctx, redisCache, *listPattern)

	case "get":
		getCmd.Parse(os.Args[2:])
		if *getKey == "" {
			fmt.Println("Необходимо указать ключ с помощью -key")
			os.Exit(1)
		}
		getValue(ctx, redisCache, *getKey)

	case "delete":
		deleteCmd.Parse(os.Args[2:])
		if *deleteKey == "" {
			fmt.Println("Необходимо указать ключ с помощью -key")
			os.Exit(1)
		}
		removeKey(ctx, redisCache, *deleteKey)

	case "flush":
		flushCmd.Parse(os.Args[2:])
		if !*flushConfirm {
			fmt.Println("Для подтверждения операции очистки используйте -confirm=true")
			os.Exit(1)
		}
		flushAll(ctx, redisCache)

	case "zrange":
		zrangeCmd.Parse(os.Args[2:])
		if *zrangeKey == "" {
			fmt.Println("Необходимо указать ключ с помощью -key")
			os.Exit(1)
		}
		getZRange(ctx, redisCache, *zrangeKey, *zrangeStart, *zrangeStop)

	default:
		fmt.Printf("Неизвестная команда %q\n", os.Args[1])
		fmt.Println("Доступные команды: list, get, delete, flush, zrange")
		os.Exit(1)
	}
}

func listKeys(ctx context.Context, c *cache.RedisCache, pattern string) {
	client := c.GetClient()
	if client == nil {
		log.Fatalf("Не удалось получить клиент Redis")
		return
	}

	var cursor uint64
	var keys []string
	var err error

	fmt.Printf("Ключи по шаблону %q:\n", pattern)

	for {
		keys, cursor, err = client.Scan(ctx, cursor, pattern, 10).Result()
		if err != nil {
			log.Fatalf("Ошибка при сканировании ключей: %v", err)
			return
		}

		for _, key := range keys {
			fmt.Println(key)
		}

		if cursor == 0 {
			break
		}
	}
}

func getValue(ctx context.Context, c *cache.RedisCache, key string) {
	value, err := c.Get(ctx, key)
	if err != nil {
		log.Fatalf("Ошибка получения значения по ключу %q: %v", key, err)
	}

	fmt.Printf("Ключ: %s\nЗначение: %s\n", key, value)
}

func removeKey(ctx context.Context, c *cache.RedisCache, key string) {
	err := c.Delete(ctx, key)
	if err != nil {
		log.Fatalf("Ошибка удаления ключа %q: %v", key, err)
	}

	fmt.Printf("Ключ %q успешно удален\n", key)
}

func flushAll(ctx context.Context, c *cache.RedisCache) {
	client := c.GetClient()
	if client == nil {
		log.Fatalf("Не удалось получить клиент Redis")
		return
	}

	_, err := client.FlushAll(ctx).Result()
	if err != nil {
		log.Fatalf("Ошибка при очистке Redis: %v", err)
		return
	}

	fmt.Println("Redis успешно очищен")
}

func getZRange(ctx context.Context, c *cache.RedisCache, key string, start, stop int64) {
	values, err := c.ZRevRangeWithScores(ctx, key, start, stop)
	if err != nil {
		log.Fatalf("Ошибка получения элементов множества %q: %v", key, err)
	}

	fmt.Printf("Элементы множества %q:\n", key)

	if len(values) == 0 {
		fmt.Println("Множество пусто или не существует")
		return
	}

	for member, score := range values {
		fmt.Printf("%.2f: %s\n", score, member)
	}
}
