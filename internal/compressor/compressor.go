// Package compressor provides video compression functionality by invoking
// external ffmpeg processes with appropriate arguments.
package compressor

import (
	"fmt"
	"os/exec"
)

// Compress invokes ffmpeg with args based on the preset.
// Parameters:
//   - input: path to input video file
//   - output: path where output will be written
//   - args: additional ffmpeg arguments, typically from presets.BuildFFArgs
//
// Returns an error if ffmpeg execution fails.
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
