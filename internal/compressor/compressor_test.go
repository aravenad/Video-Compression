// Package compressor unit tests
package compressor

import (
	"testing"
)

// TestCompress_Error ensures that Compress returns an error under expected failure scenarios,
// such as invalid arguments or missing ffmpeg binary.
func TestCompress_Error(t *testing.T) {
	// Use an invalid ffmpeg argument to provoke an error
	err := Compress("input.mp4", "output.mp4", []string{"-invalidflag"})
	if err == nil {
		t.Fatal("Expected error from Compress, got nil")
	}
}
