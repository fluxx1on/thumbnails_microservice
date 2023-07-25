package youtube_test

import (
	"testing"

	"github.com/fluxx1on/thumbnails_microservice/external/youtube"
)

func TestGetImage(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "Test #1", args: args{"https://i.ytimg.com/vi/D0St2LH158Q/maxresdefault.jpg"}, wantErr: false},
		{name: "Test #2", args: args{"https://i.ytimg.com/vi/D0St2Lr158Q/maxresdefault.jpg"}, wantErr: true},
		{name: "Test #3", args: args{"-"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := youtube.GetImage(tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && got == nil {
				t.Errorf("GetImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
