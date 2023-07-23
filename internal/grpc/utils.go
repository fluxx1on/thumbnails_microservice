package grpc

import (
	"fmt"

	"github.com/fluxx1on/thumbnails_microservice/internal/grpc/proto"
)

// GetResponseStat count sum of errors and correct Thumbnails.
// GetResponseStat used by logger to check the traffic.
func GetResponseStat(srcList ...*proto.ThumbnailResponse) string {
	var (
		thumbCounter, errorCounter int = 0, 0
	)

	for _, resp := range srcList {
		if _, is := resp.GetContent().(*proto.ThumbnailResponse_Thumbnail); is {
			thumbCounter++
		}
		if _, is := resp.GetContent().(*proto.ThumbnailResponse_Error); is {
			errorCounter++
		}
	}

	return fmt.Sprintf("successes: %d; errors: %d.", thumbCounter, errorCounter)
}
