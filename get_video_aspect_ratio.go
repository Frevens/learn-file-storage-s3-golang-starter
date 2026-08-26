package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os/exec"
)

func getVideoAspectRatio(filePath string) (string, error) {
	var output bytes.Buffer
	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)
	cmd.Stdout = &output

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to execute ffprobe: %w", err)
	}

	var probe struct {
		Streams []struct {
			Width  int `json:"width"`
			Height int `json:"height"`
		} `json:"streams"`
	}

	if err := json.Unmarshal(output.Bytes(), &probe); err != nil {
		return "", fmt.Errorf("failed to parse ffprobe output: %w", err)
	}

	for _, stream := range probe.Streams {
		if stream.Width <= 0 || stream.Height <= 0 {
			continue
		}

		ratio := float64(stream.Width) / float64(stream.Height)

		fmt.Printf(
			"FFPROBE: width=%d height=%d ratio=%f\n",
			stream.Width,
			stream.Height,
			ratio,
		)

		if math.Abs(ratio-16.0/9.0) < 0.01 {
			return "16:9", nil
		}

		if math.Abs(ratio-9.0/16.0) < 0.01 {
			return "9:16", nil
		}

		return "other", nil
	}

	return "", fmt.Errorf("ffprobe output contains no video dimensions")
}