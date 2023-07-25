package cache

import (
	"context"
	"fmt"

	"github.com/fluxx1on/thumbnails_microservice/internal/grpc/proto"
	"github.com/fluxx1on/thumbnails_microservice/libs/utils"
	"github.com/go-redis/redis"
	"golang.org/x/exp/slog"
)

const (
	baseKey = "video:"
	curDir  = "/libs/cache"

	ErrClosed = "redis: client is closed"
)

type Cache interface {
	Get(context.Context, string) *proto.ThumbnailResponse
	GetSeries(context.Context, ...string) ([]*proto.ThumbnailResponse, []string)
	SetSeries(context.Context, ...*proto.Thumbnail)
}

var _ Cache = (*RedisQuery)(nil)

type RedisQuery struct {
	Redis *redis.Client

	Context context.Context
}

func NewRedisQuery(ctx context.Context, client *redis.Client) *RedisQuery {
	return &RedisQuery{
		Redis:   client,
		Context: ctx,
	}
}

func (q *RedisQuery) Get(ctx context.Context, videoID string) *proto.ThumbnailResponse {
	var (
		hash = baseKey + videoID
		exec = q.Redis.HGetAll(hash)
	)

	if exec.Err() == nil {
		resp := utils.NewThumbnailResponse(exec)
		if resp != nil && resp.GetThumbnail().GetId() == videoID {
			slog.Debug("Searching in redis cache",
				fmt.Sprintf("%s in cache", resp.GetThumbnail().GetId()),
			)
			return resp
		}
	}

	return nil
}

func (q *RedisQuery) GetSeries(ctx context.Context, poolVideoID ...string) ([]*proto.ThumbnailResponse, []string) {
	var (
		thumbnailPool []*proto.ThumbnailResponse
		notInCache    []string = nil
		pipeline               = q.Redis.Pipeline()
	)

	for _, str := range poolVideoID {
		hash := baseKey + str
		pipeline.HGetAll(hash)
	}

	executed, err := pipeline.Exec()
	if err != nil {
		if err.Error() == ErrClosed {
			slog.Error("Redis pipeline execution failed", err, curDir)
			return nil, poolVideoID
		}
	}

	for index, ex := range executed {
		if ex.Err() == nil {
			thumbnail := utils.NewThumbnailResponse(ex.(*redis.StringStringMapCmd))
			if thumbnail != nil {
				thumbnailPool = append(thumbnailPool, thumbnail)
			} else {
				notInCache = append(notInCache, poolVideoID[index])
			}
		} else {
			notInCache = append(notInCache, poolVideoID[index])
		}
	}

	slog.Debug("Searching in redis cache",
		fmt.Sprintf("at cache: %d; new: %d;", len(thumbnailPool), len(notInCache)),
	)

	return thumbnailPool, notInCache
}

func (q *RedisQuery) Set(ctx context.Context, video *proto.Thumbnail) {
	// Unused
}

func (q *RedisQuery) SetSeries(ctx context.Context, poolVideo ...*proto.Thumbnail) {
	pipeline := q.Redis.Pipeline()
	for _, video := range poolVideo {
		hash := baseKey + video.GetId()
		pipeline.HMSet(hash, map[string]any{
			"id":           video.GetId(),
			"url":          video.GetUrl(),
			"channelTitle": video.GetChannelTitle(),
			"title":        video.GetTitle(),
			"width":        video.GetWidth(),
			"height":       video.GetHeight(),
		})
	}
	_, err := pipeline.Exec()
	if err != nil {
		slog.Warn("Set pipeline execution failed", curDir)
	} else {
		slog.Debug("Set pipeline execution succesful", curDir)
	}
}
