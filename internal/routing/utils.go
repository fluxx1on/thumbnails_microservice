package routing

import (
	"fmt"
	"net/url"
	"strings"
)

const (
	ErrInvalidURL = "Invalid video URL"
)

// GetQueryID is cutting videoID from URL.
// Also GetQueryId validate youtube video URL.
func GetQueryID(src string) (string, error) {
	url, err := url.Parse(src)
	// source is not a URL
	if err != nil {
		return ErrInvalidURL, fmt.Errorf("")
	}

	query := url.Query()
	queryID := query.Get("v")

	// query of params not contain a videoID (incorrect URL)
	if strings.TrimSpace(queryID) == "" {
		return ErrInvalidURL, fmt.Errorf("")
	}
	return queryID, nil
}
