package config

import (
	"io"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"golang.org/x/exp/slog"
)

// Logger
type Logger struct {
	Logfile   io.Writer
	LevelInfo slog.Level
}

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
	Logger           *Logger
	YouTube          *YouTubeAPI
	Redis            *RedisClient
}

// Setup needs to set .env configuration Config
func Setup() *Config {
	var (
		cfg       Config
		levelInfo slog.Level
	)

	if err := godotenv.Load(); err != nil {
		panic("Failed reading from .env")
	}

	stage := os.Getenv("STAGE")

	if stage == "dev" {
		levelInfo = slog.LevelDebug
	} else if stage == "prod" {
		levelInfo = slog.LevelInfo
	} else {
		panic("Unpredictable stage condition")
	}

	RootDir, _ := os.Getwd()

	MediaDir := os.Getenv("MEDIA_DIR")
	if err := os.MkdirAll(RootDir+MediaDir, 0755); err != nil {
		panic("Check media directory")
	}

	cfg.ServerAddress = os.Getenv("SERVER_ADDRESS")
	cfg.ListenerProtocol = os.Getenv("LISTENER_PROTOCOL")

	// Logger
	{
		Logfile, err := os.OpenFile(RootDir+os.Getenv("LOG_FILE"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			panic("Check logfile")
		}

		cfg.Logger = &Logger{
			Logfile:   Logfile,
			LevelInfo: levelInfo,
		}
	}

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
