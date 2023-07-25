package utils

import (
	"strconv"

	"github.com/fluxx1on/thumbnails_microservice/external/serial"
	"github.com/fluxx1on/thumbnails_microservice/internal/grpc/proto"
	"github.com/go-redis/redis"
)

var (
	curDir = "/libs/utils"
)

func NewErrorThumbnailResponse(url string, err string) *proto.ThumbnailResponse {
	return &proto.ThumbnailResponse{
		Content: &proto.ThumbnailResponse_Error{
			Error: &proto.ErrorResponse{
				Url:          url,
				ErrorMessage: err,
			},
		},
	}
}

func cachedThumbnailResponse(args *redis.StringStringMapCmd) *proto.ThumbnailResponse {
	values := args.Val()

	width, err1 := strconv.Atoi(values["width"])
	height, err2 := strconv.Atoi(values["height"])
	if err1 != nil || err2 != nil {
		return nil
	}

	data := ReadMediaFile(values["id"])
	if data == nil {
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
