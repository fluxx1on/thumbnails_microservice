package youtube

import (
	"fmt"
	"net/url"
	"strings"
)

// GetQueryId is cutting videoId from url #YouTube
func GetQueryId(src string) (string, error) {
	url, err := url.Parse(src)
	if err != nil {
		return "", fmt.Errorf("url encoding failed: %v", err)
	}

	query := url.Query()
	queryId := query.Get("v")

	if strings.TrimSpace(queryId) == "" {
		return "", fmt.Errorf("url don't have videoId param")
	}
	return queryId, nil
}
