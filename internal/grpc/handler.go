package grpc

import (
	"context"
	"fmt"

	"github.com/fluxx1on/thumbnails_microservice/external/youtube"
	"github.com/fluxx1on/thumbnails_microservice/internal/grpc/proto"
	"github.com/fluxx1on/thumbnails_microservice/internal/scheduler"
	"github.com/fluxx1on/thumbnails_microservice/libs/cache"
	"github.com/fluxx1on/thumbnails_microservice/libs/utils"
	"golang.org/x/exp/slog"
)

type ThumbnailFetcher interface {
	FetchThumbnail(context.Context, *proto.GetThumbnailRequest) (*proto.ThumbnailResponse, error)
	FetchThumbnailList(context.Context, *proto.ListThumbnailRequest) ([]*proto.ThumbnailResponse, error)
}

type ThumbnailFetchService struct {
	cache     *scheduler.CacheQueue
	apiClient *youtube.YouTubeAPIClient
}

func NewThumbnailFetchService(cache *scheduler.CacheQueue,
	apiClient *youtube.YouTubeAPIClient) *ThumbnailFetchService {

	return &ThumbnailFetchService{
		cache:     cache,
		apiClient: apiClient,
	}
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

	t.cache.PutQueue(thumbnailList...)
}

func (t *ThumbnailFetchService) FetchThumbnailList(ctx context.Context, reqList *proto.ListThumbnailRequest) (
	[]*proto.ThumbnailResponse, error) {
	var (
		thumbResponse = make([]*proto.ThumbnailResponse, 0, len(reqList.GetRequests()))
		cacheListId   = make([]string, 0, len(reqList.GetRequests()))
	)

	for _, value := range reqList.GetRequests() {
		id, err := youtube.GetQueryId(value.GetUrl())
		if err != nil {
			// Append errors
			thumbResponse = append(thumbResponse,
				utils.NewErrorThumbnailResponse(value.GetUrl(), fmt.Errorf("url format is incorrect")))
		} else {
			cacheListId = append(cacheListId, id)
		}
	}

	cachedThumbnails, apiListId, err := cache.GetVideoPool(t.cache.CacheClient, cacheListId...)
	if err != nil {
		slog.Error(err.Error())
	} else {
		// Append thumbnails
		thumbResponse = append(thumbResponse, cachedThumbnails...)
	}

	apiThumbnails, errListId, err := t.apiClient.GetVideoThumbnail(ctx, apiListId...)
	if err != nil {
		slog.Error(err.Error())
	} else {

		// Try to caching
		t.cacheProducer(ctx, apiThumbnails...)

		// Append thumbnails
		thumbResponse = append(thumbResponse, apiThumbnails...)

	}

	for _, url := range errListId {
		// Append errors
		thumbResponse = append(thumbResponse,
			utils.NewErrorThumbnailResponse(url, fmt.Errorf("downloading failed; video no exist")))
	}

	if len(thumbResponse) == 0 {
		return nil, fmt.Errorf("nothing to response")
	}
	return thumbResponse, nil
}

func (t *ThumbnailFetchService) FetchThumbnail(ctx context.Context, req *proto.GetThumbnailRequest) (
	*proto.ThumbnailResponse, error) {

	// It gets video id
	// Return Error response by incorrect url query parameters
	id, err := youtube.GetQueryId(req.GetUrl())
	if err != nil {
		return utils.NewErrorThumbnailResponse(req.GetUrl(), fmt.Errorf("url format is incorrect")), nil
	}

	// Return Cached response
	cachedThumbnail := cache.GetVideo(t.cache.CacheClient, id)
	if cachedThumbnail != nil {
		return cachedThumbnail, nil
	}

	// Return response from youtube API
	apiThumbnail, errListId, err := t.apiClient.GetVideoThumbnail(ctx, id)
	if err != nil {
		slog.Error(err.Error())
	} else if apiThumbnail != nil {
		// Try to caching
		t.cacheProducer(ctx, apiThumbnail...)

		return apiThumbnail[0], nil

	}

	if errListId != nil {
		return utils.NewErrorThumbnailResponse(id, fmt.Errorf("downloading failed; video no exist")), nil
	}

	return nil, fmt.Errorf("nothing to response")
}
