package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
)


func getVideoAspectRatio(filePath string) (string, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)

	output := new(bytes.Buffer)
	cmd.Stdout = output

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("couldn't run command: %w", err)
	}

	type outputParameters struct {
		Streams []struct {
			Width int `json:"width"`
			Height int `json:"height"`
		} `json:"streams"`
	}

	outputParam := outputParameters{}
	if err := json.Unmarshal(output.Bytes(), &outputParam); err != nil {
		return "", fmt.Errorf("couldn't unmarshal ratio: %w", err)
	}

	if len(outputParam.Streams) == 0 {
		return "", fmt.Errorf("no streams found in output")
	}

	width := outputParam.Streams[0].Width
	height := outputParam.Streams[0].Height

	if width == 16 * height / 9 {
		return "16:9", nil
	} else if height == 16 * width / 9 {
		return "9:16", nil
	}

	return "other", nil
}

func processVideoForFastStart(filepath string) (string, error) {
	outputFilePath := filepath + ".processing"

	cmd := exec.Command("ffmpeg", "-i", filepath, "-c", "copy", "-movflags", "faststart", "-f", "mp4", outputFilePath)

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("couldn't run command: %w", err)
	}

	return outputFilePath, nil
}

func generatePresignedURL(s3Client *s3.Client, bucket, key string, expireTime time.Duration) (string, error) {
	newS3Client := s3.NewPresignClient(s3Client)

	r, err := newS3Client.PresignGetObject(context.Background(), &s3.GetObjectInput{
		Bucket: &bucket,
		Key: &key,
	}, s3.WithPresignExpires(expireTime))
	if err != nil {
		return "", fmt.Errorf("couldn't get presign object: %w", err)
	}

	return r.URL, nil
}

func (cfg *apiConfig) dbVideoToSignedVideo(video database.Video) (database.Video, error) {
	if video.VideoURL == nil {
		return video, nil
	}
	stringSlice := strings.SplitN(*video.VideoURL, ",", 2)

	if len(stringSlice) != 2 {
		return database.Video{}, fmt.Errorf("couldn't find bucket and/or key in url")
	}

	newVideoURL, err := generatePresignedURL(cfg.s3Client, stringSlice[0], stringSlice[1], time.Minute)
	if err != nil {
		return database.Video{}, fmt.Errorf("couldn't generate url from video: %w", err)
	}

	video.VideoURL = &newVideoURL
	return video, nil
}