package grpc

import (
	"context"

	"github.com/fluxx1on/thumbnails_microservice/internal/grpc/proto"
	"github.com/fluxx1on/thumbnails_microservice/internal/routing"
	"golang.org/x/exp/slog"
)

type ThumbnailService struct {
	// Implements
	proto.UnimplementedThumbnailServiceServer

	f routing.ThumbnailFetcher
}

func NewThumbnailService(f routing.ThumbnailFetcher) *ThumbnailService {
	return &ThumbnailService{
		f: f,
	}
}

func (s *ThumbnailService) ListThumbnail(ctx context.Context, req *proto.ListThumbnailRequest) (
	*proto.ListThumbnailResponse, error) {
	resp, err := s.f.FetchThumbnailList(ctx, req)
	respList := &proto.ListThumbnailResponse{
		Thumbnails: resp,
	}

	if err != nil { // requires respList isn't nil
		slog.Info("Requested:", req.String())
		slog.Error("User didn't get any response", err)
		return nil, err
	}

	slog.Info("ListResponse succesfully sent", GetResponseStat(resp...))
	return respList, err
}

func (s *ThumbnailService) GetThumbnail(ctx context.Context, req *proto.GetThumbnailRequest) (
	*proto.ThumbnailResponse, error) {
	resp, err := s.f.FetchThumbnail(ctx, req)

	if err != nil { // requires resp isn't nil
		slog.Info("Requested:", req.String())
		slog.Error("User didn't get any response", err)
		return nil, err
	}

	slog.Info("Response sent succesfully", GetResponseStat(resp))
	return resp, err
}
