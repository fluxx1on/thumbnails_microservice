package grpc

import (
	"context"

	"github.com/fluxx1on/thumbnails_microservice/internal/grpc/proto"
	"github.com/fluxx1on/thumbnails_microservice/libs/logger/attrs"
	"golang.org/x/exp/slog"
)

type GRPCThumbnailService struct {
	// Implements
	proto.UnimplementedThumbnailServiceServer

	f ThumbnailFetcher
}

func NewGRPCThumbnailService(f ThumbnailFetcher) *GRPCThumbnailService {
	return &GRPCThumbnailService{
		f: f,
	}
}

func (s *GRPCThumbnailService) ListThumbnail(ctx context.Context, req *proto.ListThumbnailRequest) (
	*proto.ListThumbnailResponse, error) {
	resp, err := s.f.FetchThumbnailList(ctx, req)
	respList := &proto.ListThumbnailResponse{
		Thumbnails: resp,
	}

	if err != nil { // requires respList isn't nil
		slog.Info("Requested:", attrs.Any(req.String()))
		slog.Error("User have no respond", attrs.Err(err))
		return nil, err
	}

	slog.Info("Message succesfully sent", attrs.Any(GetResponseStat(resp...)))
	return respList, err
}

func (s *GRPCThumbnailService) GetThumbnail(ctx context.Context, req *proto.GetThumbnailRequest) (
	*proto.ThumbnailResponse, error) {
	resp, err := s.f.FetchThumbnail(ctx, req)

	if err != nil { // requires resp isn't nil
		slog.Info("Requested:", attrs.Any(req.String()))
		slog.Error("User have no respond", attrs.Err(err))
		return nil, err
	}

	slog.Info("Message sent succesfully", attrs.Any(GetResponseStat(resp)))
	return resp, err
}
