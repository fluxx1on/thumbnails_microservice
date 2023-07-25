package youtube

import (
	"fmt"
	"io"
	"net/http"

	"github.com/fluxx1on/thumbnails_microservice/external/serial"
)

func GetImage(url string) (serial.ThumbnailData, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if response.StatusCode == 200 {
		body, err := io.ReadAll(response.Body)
		return body, err
	}
	return nil, fmt.Errorf("status code: %d", response.StatusCode)
}
