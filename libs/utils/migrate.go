package utils

import (
	"strconv"

	"github.com/fluxx1on/thumbnails_microservice/external/serial"
	"github.com/fluxx1on/thumbnails_microservice/internal/grpc/proto"
	"github.com/go-redis/redis"
	"golang.org/x/exp/slog"
)

var (
	curDir = "/libs/utils"
)

func NewErrorThumbnailResponse(url string, err error) *proto.ThumbnailResponse {
	return &proto.ThumbnailResponse{
		Content: &proto.ThumbnailResponse_Error{
			Error: &proto.ErrorResponse{
				Url:          url,
				ErrorMessage: err.Error(),
			},
		},
	}
}

func cachedThumbnailResponse(args *redis.StringStringMapCmd) *proto.ThumbnailResponse {
	values := args.Val()

	width, _ := strconv.Atoi(values["width"])
	height, _ := strconv.Atoi(values["height"])

	data, err := ReadMediaFile(values["id"])
	if err != nil {
		slog.Debug(err.Error())
		return nil
	}

	thumbnailResponse := &proto.ThumbnailResponse{
		Content: &proto.ThumbnailResponse_Thumbnail{
			Thumbnail: &proto.Thumbnail{
				Id:           values["id"],
				Url:          values["url"],
				ChannelTitle: values["channelTitle"],
				Title:        values["title"],
				Width:        int32(width),
				Height:       int32(height),
				File:         data,
			},
		},
	}

	return thumbnailResponse
}

func requestedThumbnailResponse(video *serial.Video) *proto.ThumbnailResponse {
	item := video.I

	thumbnailResponse := &proto.ThumbnailResponse{
		Content: &proto.ThumbnailResponse_Thumbnail{
			Thumbnail: &proto.Thumbnail{
				Id:           item.GetId(),
				Url:          item.GetUrl(),
				ChannelTitle: item.GetChannelTitle(),
				Title:        item.GetTitle(),
				Width:        item.GetWidth(),
				Height:       item.GetHeight(),
				File:         video.GetData(),
			},
		},
	}

	return thumbnailResponse
}

func NewThumbnailResponse(cmd interface{}) *proto.ThumbnailResponse {
	switch args := cmd.(type) {
	case *redis.StringStringMapCmd:
		return cachedThumbnailResponse(args)
	case *serial.Video:
		return requestedThumbnailResponse(args)
	default:
		return nil
	}
}
