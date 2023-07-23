package cache

import (
	"errors"
	"fmt"

	"github.com/fluxx1on/thumbnails_microservice/internal/grpc/proto"
	"github.com/fluxx1on/thumbnails_microservice/libs/logger/attrs"
	"github.com/fluxx1on/thumbnails_microservice/libs/utils"
	"github.com/go-redis/redis"
	"golang.org/x/exp/slog"
)

const (
	baseKey = "video:"
	curDir  = "/libs/cache"
)

var (
	ErrClosed = errors.New("redis: client is closed")
)

func GetVideo(Redis *redis.Client, videoId string) *proto.ThumbnailResponse {
	var (
		concat = baseKey + videoId
		exec   = Redis.HGetAll(concat)
	)

	if exec.Err() == nil {
		resp := utils.NewThumbnailResponse(exec)
		if resp != nil && resp.GetThumbnail().GetId() == videoId {
			return resp
		}
	}
	return nil
}

func SetVideoPool(Redis *redis.Client, poolVideo ...*proto.Thumbnail) {
	pipeline := Redis.Pipeline()
	for _, video := range poolVideo {
		pipeline.HMSet(baseKey+video.Id, map[string]any{
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
		slog.Warn("Set pipeline execution failed", attrs.Any(curDir))
	} else {
		slog.Debug("Set pipeline execution succesful", attrs.Any(curDir))
	}
}

func GetVideoPool(Redis *redis.Client, poolVideoId ...string) ([]*proto.ThumbnailResponse, []string, error) {
	var (
		thumbnailPool []*proto.ThumbnailResponse
		notInCache    []string
		pipeline      = Redis.Pipeline()
	)

	for _, str := range poolVideoId {
		concat := baseKey + str
		pipeline.HGetAll(concat)
	}

	executed, err := pipeline.Exec()
	if err != nil {
		if err.Error() == ErrClosed.Error() {
			return nil, poolVideoId, fmt.Errorf("redis pipeline execution failed: %w ... %s", err, curDir)
		}
	}

	for index, ex := range executed {
		if ex.Err() == nil {
			thumbnail := utils.NewThumbnailResponse(ex.Args())
			if thumbnail != nil {
				thumbnailPool = append(thumbnailPool, thumbnail)
			} else {
				notInCache = append(notInCache, poolVideoId[index])
			}
		} else {
			notInCache = append(notInCache, poolVideoId[index])
		}
	}

	return thumbnailPool, notInCache, nil
}

// func GetVideosKeys(Redis *redis.Client) ([]string, error) {
// 	return Redis.Keys(baseKey).Result()
// }
