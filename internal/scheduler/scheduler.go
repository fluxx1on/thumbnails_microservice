package scheduler

import (
	"context"

	"github.com/fluxx1on/thumbnails_microservice/internal/grpc/proto"
	"github.com/fluxx1on/thumbnails_microservice/libs/cache"
	"github.com/fluxx1on/thumbnails_microservice/libs/logger/attrs"
	"github.com/fluxx1on/thumbnails_microservice/libs/utils"
	"github.com/go-redis/redis"
	"golang.org/x/exp/slog"
)

var (
	curDir = "/internal/scheduler"
)

type CacheQueue struct {
	CacheClient *redis.Client

	// Queue is like task queue/schedule from broker
	queue chan []*proto.Thumbnail

	// Context to stop broker queue before connection will be lost
	// That make possible transact cache to redis without loss
	ctx context.Context

	ctxCancel context.CancelFunc
}

func NewCacheQueue(ctx context.Context, Redis *redis.Client) *CacheQueue {
	ctx, cancel := context.WithCancel(ctx)

	return &CacheQueue{
		CacheClient: Redis,
		queue:       make(chan []*proto.Thumbnail, 100),
		ctx:         ctx,
		ctxCancel:   cancel,
	}
}

// putCache inspect that slice doesn't contain same video multiple times.
// So after inspection putCache provide video meta data to redis
// and thumbnail image to filesystem.
func (q *CacheQueue) putCache(thumbResp []*proto.Thumbnail) {
	var (
		thumbList = make([]*proto.Thumbnail, 0, len(thumbResp))
		dict      = make(map[string]int, len(thumbResp))
	)

	for index, thumb := range thumbResp {
		dict[thumb.GetId()] = index
	}

	for key, val := range dict {
		if err := utils.WriteMediaFile(thumbResp[val].GetFile(), key); err != nil {
			slog.Warn("Writing image file denied", attrs.Err(err), attrs.Any(curDir))
		}
		thumbList = append(thumbList, thumbResp[val])
	}

	cache.SetVideoPool(q.CacheClient, thumbList...)

	slog.Debug("All files were written succesfully")
}

func (q *CacheQueue) PutQueue(thumb ...*proto.Thumbnail) {
	q.queue <- thumb
}

// JobRunning is One-Thread consumer.
// It reads CacheQueue.queue and send it to CacheQueue.putCache.
// Can be shutted down by context Cancelation.
func (q *CacheQueue) JobRunning() {
	for {
		select {
		case thumb := <-q.queue:
			q.putCache(thumb)
		case <-q.ctx.Done():
			return
		}
	}
}

// ShutdownJob cancel context and stop JobRunning safety.
func (q *CacheQueue) ShutdownJob() {
	q.ctxCancel()
}
