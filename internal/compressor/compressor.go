package compressor

import (
	"fmt"
	"os/exec"
)

// Compress invokes ffmpeg with args based on the preset.
// For now, we’ll just print a stub.
func Compress(input, output, preset string) error {
	// TODO: map preset → []string of ffmpeg args (e.g. "-c:v libx264", ...)
	args := []string{
		"-i", input,
		// placeholder: override output
		output,
	}

	fmt.Printf("Running: ffmpeg %v\n", args)
	cmd := exec.Command("ffmpeg", args...)
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg error: %w", err)
	}
	return nil
}
