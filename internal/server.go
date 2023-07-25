package internal

import (
	"context"
	"net"

	"golang.org/x/exp/slog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	// Current module
	"github.com/fluxx1on/thumbnails_microservice/cmd/config"
	"github.com/fluxx1on/thumbnails_microservice/external/youtube"
	"github.com/fluxx1on/thumbnails_microservice/internal/cache"
	igrpc "github.com/fluxx1on/thumbnails_microservice/internal/grpc"
	"github.com/fluxx1on/thumbnails_microservice/internal/grpc/proto"
	"github.com/fluxx1on/thumbnails_microservice/internal/routing"
	"github.com/fluxx1on/thumbnails_microservice/internal/scheduler"
	"github.com/go-redis/redis"
)

type GRPC struct {
	listener  net.Listener
	server    *grpc.Server
	scheduler *scheduler.CacheQueue
}

func (g *GRPC) StartUp(cfg *config.Config, RedisConn *redis.Client) {
	// Listener starting
	var err error
	g.listener, err = net.Listen(cfg.ListenerProtocol, cfg.ServerAddress)
	if err != nil {
		slog.Error("Failed to listen", err)
	}

	slog.Info(g.listener.Addr().String())

	// gRPC creating
	g.server = grpc.NewServer()
	reflection.Register(g.server)

	// RedisQuery caching setup
	CacheClient := cache.NewRedisQuery(context.Background(), RedisConn)

	// Scheduler setup
	CacheScheduler := scheduler.NewCacheQueue(context.Background(), CacheClient)
	g.scheduler = CacheScheduler

	// GRPCThumbnailService setup
	uAPI := youtube.NewAPIClient(cfg.YouTube) // YouTubeAPI init
	srv := igrpc.NewThumbnailService(routing.NewThumbnailFetchService(
		CacheScheduler, uAPI,
	))
	proto.RegisterThumbnailServiceServer(g.server, srv)

	// Server starting
	go func() {
		if err := g.server.Serve(g.listener); err != nil {
			slog.Error("Failed to serve: %v", err)
		}
	}()

	// Start consumer
	go g.scheduler.JobRunning()

	slog.Info("gRPC server started on address:", cfg.ServerAddress)
}

func (g *GRPC) Stop() {
	if g.scheduler != nil {
		g.scheduler.ShutdownJob()
	}
	g.server.Stop()
	g.listener.Close()
}
