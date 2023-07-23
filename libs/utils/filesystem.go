package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"

	"github.com/fluxx1on/thumbnails_microservice/external/serial"
)

var (
	rootDir, _ = os.Getwd()
	mediaDir   = rootDir + "/media/"
)

func HashString(str string) [32]byte {
	return sha256.Sum256([]byte(str))
}

func GetHashDirName(id string) string {
	hash := HashString(id)
	hashString := hex.EncodeToString(hash[:])
	return string(hashString[0])
}

func GetFilePath(videoId string) string {
	filename := HashString(videoId)
	hashString := hex.EncodeToString(filename[:])
	hashdir := GetHashDirName(videoId) + "/"
	return fmt.Sprintf("%s.jpg", mediaDir+hashdir+hashString)
}

func ReadMediaFile(videoId string) ([]byte, error) {
	file, err := os.Open(GetFilePath(videoId))
	if err != nil {
		return nil, fmt.Errorf("nothing to read; file not exist ... %s", curDir)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	return data, fmt.Errorf("reading closed: %w ... %s", err, curDir)
}

func WriteMediaFile(imageData serial.ThumbnailData, videoId string) error {

	// Creating directory if no exist
	err := os.MkdirAll(mediaDir+GetHashDirName(videoId), 0755)
	if err != nil {
		return fmt.Errorf("directory unreached: %w", err)
	}

	// Creating and opening new file
	file, err := os.Create(GetFilePath(videoId))
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
