package youtube

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/fluxx1on/thumbnails_microservice/cmd/config"
	"github.com/fluxx1on/thumbnails_microservice/external/serial"
	"github.com/fluxx1on/thumbnails_microservice/internal/grpc/proto"
	"github.com/fluxx1on/thumbnails_microservice/libs/utils"
)

var (
	curDir        = "/external/youtube"
	timeout       = 15 * time.Second
	baseVideosUrl = "https://youtube.googleapis.com/youtube/v3/videos?part=snippet"
)

type YouTubeAPIClient struct {
	httpClient *http.Client
	cfg        *config.YouTubeAPI
}

func NewYouTubeAPIClient(YouTubeCfg *config.YouTubeAPI) *YouTubeAPIClient {
	httpClient := &http.Client{}

	return &YouTubeAPIClient{
		cfg:        YouTubeCfg,
		httpClient: httpClient,
	}
}

func (y *YouTubeAPIClient) getUrl(videoId ...string) string {
	builder := &strings.Builder{}

	builder.WriteString(baseVideosUrl)
	for _, str := range videoId {
		builder.WriteString("&id=" + str)
	}

	return builder.String() + "&key=" + y.cfg.APIKey
}

func (y *YouTubeAPIClient) getVideos(videoId ...string) (*serial.ListVideoSerializer, error) {
	// Make request
	req, err := http.NewRequest("GET", y.getUrl(videoId...), nil)
	if err != nil {
		return nil, err
	}

	// Headers
	req.Header.Set("Accept", "application/json")

	// Get response
	resp, err := y.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("youtube no respond: %w", err)
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("youtube request failed: %d ... %s", resp.StatusCode, curDir)
	}

	// Deserializing response.Body
	var videos serial.ListVideoSerializer
	err = json.NewDecoder(resp.Body).Decode(&videos)
	if err != nil || videos.Items == nil {
		return nil, fmt.Errorf("error while decoding: %w / %s", err, y.getUrl(videoId...))
	}

	return &videos, nil
}

func (y *YouTubeAPIClient) getThumbnails(ctx context.Context, imageUrl ...string) ([]serial.ThumbnailData, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var (
		wg            sync.WaitGroup
		thumbnailList = make([]serial.ThumbnailData, 0, len(imageUrl))
	)

	for _, url := range imageUrl {
		wg.Add(1)
		go func(ctx context.Context, url string) {
			defer wg.Done()

			thumbnail, _ := GetImage(url)
			select {
			case <-ctx.Done():
				return
			default:
				thumbnailList = append(thumbnailList, thumbnail)
			}
		}(ctx, url)
	}

	wg.Wait()

	// Some thumbnails didn't download. Error needs to be provided to user.
	if ctx.Err() != nil {
		return thumbnailList, fmt.Errorf("request timeout")
	}

	// Clear downloading
	return thumbnailList, nil
}

// GetVideoThumbnail gets videos meta data and thumbnails
func (y *YouTubeAPIClient) GetVideoThumbnail(ctx context.Context, videoId ...string) (
	[]*proto.ThumbnailResponse, []string, error) {
	var (
		errListId             []string
		thumbnailResponseList []*proto.ThumbnailResponse
	)

	videos, err := y.getVideos(videoId...)
	if err != nil { // requires that videos not nil
		return nil, videoId, err
	}

	var imageUrls = make([]string, 0, len(videos.Items))
	for _, item := range videos.Items {
		imageUrls = append(imageUrls, item.GetUrl())
	}

	thumbnails, err := y.getThumbnails(ctx, imageUrls...)
	if err != nil {
		return nil, videoId, err
	}

	for i := range videoId {
		if i < len(videos.Items) && thumbnails[i] != nil {
			video := &serial.Video{
				I:    &videos.Items[i],
				Data: thumbnails[i],
			}
			newThumbnail := utils.NewThumbnailResponse(video)
			thumbnailResponseList = append(thumbnailResponseList, newThumbnail)
		} else {
			errListId = append(errListId, videoId[i])
		}
	}

	return thumbnailResponseList, errListId, nil
}

// func (y *YouTubeAPIClient) GetVideoThumbnailParralel(ctx context.Context, poolVideoId ...string) ([]*proto.ThumbnailResponse, error) {
// 	ctx, cancel := context.WithTimeout(ctx, timeout)
// 	defer cancel()

// 	var (
// 		wg            sync.WaitGroup
// 		thumbnailList = make([]*proto.ThumbnailResponse, 0, len(poolVideoId))
// 	)

// 	for _, videoId := range poolVideoId {
// 		wg.Add(1)
// 		go func(ctx context.Context, videoId string) {
// 			defer wg.Done()

// 			thumbnailResponse := y.GetVideoThumbnail(ctx, videoId)
// 			select {
// 			case <-ctx.Done():
// 				return
// 			default:
// 				thumbnailList = append(thumbnailList, thumbnailResponse)
// 			}
// 		}(ctx, videoId)
// 	}

// 	wg.Wait()

// 	return thumbnailList, nil
// }
