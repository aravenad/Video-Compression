// This file contains utility functions for the compress command.
package main

import (
	"os"
	"path/filepath"
	"strings"
)

// deriveOutput returns the output file path for a given input file.
// Behavior:
//  1. If output == "": place alongside the input file, appending "-compressed" to the basename.
//  2. If output exists and is a directory, place the file inside it, keeping the same basename.
//  3. If output ends with a path separator, treat it as a directory.
//  4. Otherwise, treat output as the explicit file path.
//
// Examples:
//   - deriveOutput("video.mp4", "") -> "video-compressed.mp4" (in same directory)
//   - deriveOutput("video.mp4", "out/") -> "out/video.mp4"
//   - deriveOutput("video.mp4", "renamed.mp4") -> "renamed.mp4"
func deriveOutput(infile, output string) string {
	if output == "" {
		dir := filepath.Dir(infile)
		ext := filepath.Ext(infile)
		name := strings.TrimSuffix(filepath.Base(infile), ext)
		return filepath.Join(dir, name+"-compressed"+ext)
	}
	if info, err := os.Stat(output); err == nil && info.IsDir() {
		return filepath.Join(output, filepath.Base(infile))
	}
	if strings.HasSuffix(output, string(os.PathSeparator)) || strings.HasSuffix(output, "/") {
		return filepath.Join(output, filepath.Base(infile))
	}
	return output
}
