package grpc

import (
	"fmt"

	"github.com/fluxx1on/thumbnails_microservice/internal/grpc/proto"
)

// GetResponseStat count sum of errors and correct Thumbnails.
// Used by gRPC service to check and log the traffic.
func GetResponseStat(srcList ...*proto.ThumbnailResponse) string {
	var (
		thumbCounter, errorCounter int = 0, 0
	)

	for _, resp := range srcList {
		if thumb := resp.GetThumbnail(); thumb != nil {
			thumbCounter++
		} else if err := resp.GetError(); err != nil {
			errorCounter++
		}
	}

	return fmt.Sprintf("successes: %d; errors: %d.", thumbCounter, errorCounter)
}
