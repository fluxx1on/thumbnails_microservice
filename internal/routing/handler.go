package routing

import (
	"context"
	"fmt"

	"github.com/fluxx1on/thumbnails_microservice/external/youtube"
	"github.com/fluxx1on/thumbnails_microservice/internal/cache"
	"github.com/fluxx1on/thumbnails_microservice/internal/grpc/proto"
	"github.com/fluxx1on/thumbnails_microservice/internal/scheduler"
	"github.com/fluxx1on/thumbnails_microservice/libs/utils"
	"golang.org/x/exp/slog"
)

const (
	ErrDownloadVideo = "Downloading failed; video no exist"
)

type ThumbnailFetcher interface {
	FetchThumbnail(context.Context, *proto.GetThumbnailRequest) (*proto.ThumbnailResponse, error)
	FetchThumbnailList(context.Context, *proto.ListThumbnailRequest) ([]*proto.ThumbnailResponse, error)
}

var _ ThumbnailFetcher = (*ThumbnailFetchService)(nil)

type ThumbnailFetchService struct {
	cacheQ    *scheduler.CacheQueue
	apiClient *youtube.APIClient
}

func NewThumbnailFetchService(cache *scheduler.CacheQueue,
	apiClient *youtube.APIClient) *ThumbnailFetchService {

	return &ThumbnailFetchService{
		cacheQ:    cache,
		apiClient: apiClient,
	}
}

func (t *ThumbnailFetchService) getCacheClient() *cache.RedisQuery {
	return t.cacheQ.GetCacheClient()
}

// cacheProducer is a producer for CacheQueue
func (t *ThumbnailFetchService) cacheProducer(ctx context.Context, videoList ...*proto.ThumbnailResponse) {
	var thumbnailList = make([]*proto.Thumbnail, 0, len(videoList))

	// Require that all responses are thumbnails, not errors
	for _, resp := range videoList {
		if resp.GetError() != nil {
			slog.Warn("Try to caching requested with errors")
			return
		}
		thumbnailList = append(thumbnailList, resp.GetThumbnail())
	}

	t.cacheQ.PutQueue(thumbnailList...)
}

// FetchThumbnailList is intermediate node that gather all Thumbnails from cache or API
func (t *ThumbnailFetchService) FetchThumbnailList(ctx context.Context, reqList *proto.ListThumbnailRequest) (
	[]*proto.ThumbnailResponse, error) {
	var (
		thumbResponse = make([]*proto.ThumbnailResponse, 0, len(reqList.GetRequests()))
		cacheListID   = make([]string, 0, len(reqList.GetRequests()))
	)

	// Validate requested URLs
	// By Error append ErrorResponse
	for _, value := range reqList.GetRequests() {
		id, err := GetQueryID(value.GetUrl())
		if err != nil {
			thumbResponse = append(thumbResponse,
				utils.NewErrorThumbnailResponse(value.GetUrl(), id))
		} else {
			cacheListID = append(cacheListID, id)
		}
	}

	// Append cached ThumbnailReponses from Redis and filesystem
	cachedThumbnails, apiListID := t.getCacheClient().GetSeries(ctx, cacheListID...)
	thumbResponse = append(thumbResponse, cachedThumbnails...)

	// Append ThumbnailResponses from youtube API
	apiThumbnails, errListID := t.apiClient.GetVideoThumbnail(ctx, apiListID...)
	if apiThumbnails != nil {
		// Try to caching
		t.cacheProducer(ctx, apiThumbnails...)

		thumbResponse = append(thumbResponse, apiThumbnails...)
	}

	// Incorrect Video IDs; Append ErrorResponses
	for _, url := range errListID {
		thumbResponse = append(thumbResponse,
			utils.NewErrorThumbnailResponse(url, ErrDownloadVideo))
	}

	if len(thumbResponse) == 0 {
		return nil, fmt.Errorf("nothing to response")
	}
	return thumbResponse, nil
}

// FetchThumbnail is intermediate node that gather all Thumbnails from cache or API
func (t *ThumbnailFetchService) FetchThumbnail(ctx context.Context, req *proto.GetThumbnailRequest) (
	*proto.ThumbnailResponse, error) {

	// It gets video ID
	// Return ErrorResponse by incorrect url query parameters
	id, err := GetQueryID(req.GetUrl())
	if err != nil {
		return utils.NewErrorThumbnailResponse(req.GetUrl(), id), nil
	}

	// Return Cached response
	cachedThumbnail := t.getCacheClient().Get(ctx, id)
	if cachedThumbnail != nil {
		return cachedThumbnail, nil
	}

	// Return ThumbnailResponse from youtube API
	apiThumbnail, errListID := t.apiClient.GetVideoThumbnail(ctx, id)
	if apiThumbnail != nil {
		// Try to caching
		t.cacheProducer(ctx, apiThumbnail...)

		return apiThumbnail[0], nil
	}

	// Nothing finded; Return ErrorResponse
	if errListID != nil {
		return utils.NewErrorThumbnailResponse(id, ErrDownloadVideo), nil
	}

	return nil, fmt.Errorf("nothing to response")
}
