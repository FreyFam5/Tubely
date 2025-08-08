package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
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