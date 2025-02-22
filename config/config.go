package config

type Config struct {
	ServerAddress string
	DBType        string
	DBSource      string
	RedisAddr     string
	RedisPassword string
	RedisDB       int
}

func LoadConfig() Config {
	return Config{
		ServerAddress: ":8080",
		DBType:        "sqlite",
		DBSource:      "manga.db",
		RedisAddr:     "localhost:6379",
		RedisPassword: "",
		RedisDB:       1,
	}
}
