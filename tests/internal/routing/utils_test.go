package routing_test

import (
	"testing"

	"github.com/fluxx1on/thumbnails_microservice/internal/routing"
)

const (
	ErrInvalidURL = "Invalid video URL"
)

func TestGetQueryID(t *testing.T) {
	type args struct {
		src string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{name: "Test #1", args: args{"https://www.youtube.com/watch?v=Gmlh0NrvzP0&ab_channel=AnthonyGG"}, want: "Gmlh0NrvzP0", wantErr: false},
		{name: "Test #2", args: args{"https://www.youtube.com/watch?k=Gmlh0NrvzP9&ab_channel=AnthonyGG"}, want: ErrInvalidURL, wantErr: true},
		{name: "Test #3", args: args{"Negative case"}, want: ErrInvalidURL, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := routing.GetQueryID(tt.args.src)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetQueryId() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetQueryId() = %v, want %v", got, tt.want)
			}
		})
	}
}
