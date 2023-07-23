package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// YoutubeAPI store secret keys and provide it for YoutubeAPIClient
type YouTubeAPI struct {
	APIKey string
}

// RedisClient configuration settings for Redis
type RedisClient struct {
	Address  string
	DB       int
	PoolSize int
}

// Config is a configuration struct that store enviromental variables
type Config struct {
	ServerAddress    string
	ListenerProtocol string
	YouTube          *YouTubeAPI
	Redis            *RedisClient
}

// Setup needs to set .env configuration Config
func Setup() *Config {
	if err := godotenv.Load(); err != nil {
		panic("Ошибка чтения из .env")
	}

	RootDir, _ := os.Getwd()
	MediaDir := os.Getenv("MEDIA_DIR")
	if err := os.MkdirAll(RootDir+MediaDir, 0755); err != nil {
		panic("Check media directory")
	}

	var cfg Config

	cfg.ServerAddress = os.Getenv("SERVER_ADDRESS")
	cfg.ListenerProtocol = os.Getenv("LISTENER_PROTOCOL")

	// YouTube
	{
		cfg.YouTube = &YouTubeAPI{
			APIKey: os.Getenv("YOUTUBE_APIKEY"),
		}
	}

	// Redis
	{
		poolSize, _ := strconv.Atoi(os.Getenv("REDIS_CONNECTION_POOL"))
		db, _ := strconv.Atoi(os.Getenv("REDIS_DB"))

		cfg.Redis = &RedisClient{
			Address:  os.Getenv("REDIS_ADDRESS"),
			DB:       db,
			PoolSize: poolSize,
		}
	}

	return &cfg
}
