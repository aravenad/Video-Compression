package compressor

import (
	"fmt"
	"os/exec"
)

// Compress invokes ffmpeg with args based on the preset.
func Compress(input, output string, args []string) error {
	fullArgs := []string{"-i", input}
	fullArgs = append(fullArgs, args...)
	fullArgs = append(fullArgs, output)

	fmt.Printf("Running: ffmpeg %v\n", fullArgs)
	cmd := exec.Command("ffmpeg", fullArgs...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}
