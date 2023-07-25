package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"

	"github.com/fluxx1on/thumbnails_microservice/external/serial"
	"golang.org/x/exp/slog"
)

var (
	rootDir, _ = os.Getwd()
	mediaDir   = rootDir + "/media/"
)

func hashString(str string) [32]byte {
	return sha256.Sum256([]byte(str))
}

func getHashDirName(id string) string {
	hash := hashString(id)
	hashString := hex.EncodeToString(hash[:])
	return string(hashString[0])
}

func getFilePath(videoID string) string {
	filename := hashString(videoID)
	hashString := hex.EncodeToString(filename[:])
	hashdir := getHashDirName(videoID) + "/"
	return fmt.Sprintf("%s.jpg", mediaDir+hashdir+hashString)
}

func ReadMediaFile(videoID string) []byte {
	file, err := os.Open(getFilePath(videoID))
	if err != nil {
		slog.Debug("nothing to read; file not exist", curDir)
		return nil
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		slog.Warn("reading closed", err, curDir)
	}
	return data
}

func WriteMediaFile(imageData serial.ThumbnailData, videoID string) error {

	// Creating directory if no exist
	err := os.MkdirAll(mediaDir+getHashDirName(videoID), 0755)
	if err != nil {
		return fmt.Errorf("directory unreached: %w", err)
	}

	// Creating and opening new file
	file, err := os.Create(getFilePath(videoID))
	if err != nil {
		return err
	}
	defer file.Close()

	// Write image bytes to file
	if length, err := file.Write(imageData); err != nil {
		return fmt.Errorf("file didn't write correctly: %d / %d", length, len(imageData))
	}
	return nil
}
