package youtube

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/fluxx1on/thumbnails_microservice/cmd/config"
	"github.com/fluxx1on/thumbnails_microservice/external/serial"
	"github.com/fluxx1on/thumbnails_microservice/internal/grpc/proto"
	"github.com/fluxx1on/thumbnails_microservice/libs/utils"
	"golang.org/x/exp/slog"
)

var (
	curDir        = "/external/youtube"
	timeout       = 15 * time.Second
	baseVideosURL = "https://youtube.googleapis.com/youtube/v3/videos?part=snippet"
)

type API interface {
	GetVideos(...string) *serial.ListVideoSerializer
	GetThumbnails(context.Context, ...string) []serial.ThumbnailData
	GetVideoThumbnail(context.Context, ...string) ([]*proto.ThumbnailResponse, []string)
}

var _ API = (*APIClient)(nil)

type APIClient struct {
	httpClient *http.Client
	cfg        *config.YouTubeAPI
}

func NewAPIClient(YouTubeCfg *config.YouTubeAPI) *APIClient {
	httpClient := &http.Client{}

	return &APIClient{
		cfg:        YouTubeCfg,
		httpClient: httpClient,
	}
}

func (y *APIClient) GetURL(videoID ...string) string {
	builder := &strings.Builder{}

	builder.WriteString(baseVideosURL)
	for _, str := range videoID {
		builder.WriteString("&id=" + str)
	}

	return builder.String() + "&key=" + y.cfg.APIKey
}

func (y *APIClient) GetVideos(videoID ...string) *serial.ListVideoSerializer {
	var URL = y.GetURL(videoID...)

	// Make request
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		slog.Error("Unknown", err, curDir)
		return nil
	}

	// Headers
	req.Header.Set("Accept", "application/json")

	// Get response
	resp, err := y.httpClient.Do(req)
	if err != nil {
		slog.Error("YouTube no respond", err, curDir)
		return nil
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		slog.Error("YouTube request failed", resp.StatusCode, curDir)
		return nil
	}

	// Deserializing response.Body
	var videos serial.ListVideoSerializer
	err = json.NewDecoder(resp.Body).Decode(&videos)
	if err != nil || videos.IsEmpty() {
		slog.Debug("Errors while decoding", URL)
		return nil
	}

	return &videos
}

func (y *APIClient) GetThumbnails(ctx context.Context, imageURL ...string) []serial.ThumbnailData {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var (
		wg            sync.WaitGroup
		thumbnailList = make([]serial.ThumbnailData, 0, len(imageURL))
	)

	for _, url := range imageURL {
		wg.Add(1)
		go func(ctx context.Context, url string) {
			defer wg.Done()

			thumbnail, err := GetImage(url)
			if err != nil { // requires that thumbnail is nil
				slog.Debug("Bad response from YT API", err)
			}
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
		slog.Error("Request timeout", ctx.Err(), curDir)
		return nil
	}

	// Clear downloading
	return thumbnailList
}

// GetVideoThumbnail gets videos meta data and thumbnails
func (y *APIClient) GetVideoThumbnail(ctx context.Context, videoID ...string) (
	[]*proto.ThumbnailResponse, []string) {
	if len(videoID) == 0 {
		return nil, nil
	}

	var (
		errListID             []string
		thumbnailResponseList []*proto.ThumbnailResponse
	)

	videos := y.GetVideos(videoID...)
	if videos == nil { // requires that videos are not nil
		return nil, videoID
	}

	imageUrls := make([]string, 0, len(videos.Items))
	for _, item := range videos.Items {
		imageUrls = append(imageUrls, item.GetUrl())
	}

	thumbnails := y.GetThumbnails(ctx, imageUrls...)
	if thumbnails == nil { // requires that thumbnails are not nil
		return nil, videoID
	}

	for i := range videoID {
		if i < len(videos.Items) && thumbnails[i] != nil {
			video := &serial.Video{
				I:    &videos.Items[i],
				Data: thumbnails[i],
			}
			newThumbnail := utils.NewThumbnailResponse(video)
			thumbnailResponseList = append(thumbnailResponseList, newThumbnail)
		} else {
			errListID = append(errListID, videoID[i])
		}
	}

	return thumbnailResponseList, errListID
}
