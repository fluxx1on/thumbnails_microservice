package youtube

import (
	"io"
	"net/http"

	"github.com/fluxx1on/thumbnails_microservice/external/serial"
)

func GetImage(url string) (serial.ThumbnailData, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(response.Body)
	return body, err
}
