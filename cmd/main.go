package main

import (
	"context"
	baseLog "log"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-redis/redis"
	"golang.org/x/exp/slog"

	// Current module
	"github.com/fluxx1on/thumbnails_microservice/cmd/config"
	"github.com/fluxx1on/thumbnails_microservice/internal"
	"github.com/fluxx1on/thumbnails_microservice/libs/logger/attrs"
	"github.com/fluxx1on/thumbnails_microservice/libs/logger/handler"
)

func main() {
	signalCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Config
	cfg := config.Setup()

	// Logger
	log := slog.New(handler.NewColorfulHandler(baseLog.Default().Writer()))
	slog.SetDefault(log)

	// Redis
	redis := redis.NewClient(
		&redis.Options{
			Addr:     cfg.Redis.Address,
			DB:       cfg.Redis.DB,
			PoolSize: cfg.Redis.PoolSize,
		})
	if _, err := redis.Ping().Result(); err != nil {
		log.Error("Redis don't ping", err)
		return
	}

	// gRPC Server starting
	server := &internal.GRPC{}
	server.StartUp(cfg, redis)

	<-signalCtx.Done()

	// Shutting down
	log.Info("server shutting down. all connection will be terminated")

	finished := make(chan struct{}, 1)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go func() {
		server.Stop()
		redis.Close()
		finished <- struct{}{}
	}()

	select {
	case <-shutdownCtx.Done():
		log.Error("server shutdown:", attrs.Err(signalCtx.Err(), shutdownCtx.Err()))
	case <-finished:
		log.Info("succesfully finished")
	}

}
