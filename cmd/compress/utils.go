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
	// Normalize input path to platform-specific format
	infile = filepath.Clean(infile)

	// Case 1: No output specified
	if output == "" {
		dir := filepath.Dir(infile)
		ext := filepath.Ext(infile)
		name := strings.TrimSuffix(filepath.Base(infile), ext)
		return filepath.Clean(filepath.Join(dir, name+"-compressed"+ext))
	}

	// Save original output before cleaning for separator check
	originalOutput := output

	// Normalize output path
	output = filepath.Clean(output)

	// Case 2: Output is an existing directory
	if info, err := os.Stat(output); err == nil && info.IsDir() {
		return filepath.Clean(filepath.Join(output, filepath.Base(infile)))
	}

	// Case 3: Output ends with a separator (intended as directory)
	// Use the original output string to check for path separators
	if endsWithSeparator(originalOutput) {
		return filepath.Clean(filepath.Join(output, filepath.Base(infile)))
	}

	// Case 4: Output is an explicit file path
	return output
}

// endsWithSeparator checks if the path ends with a separator
// This handles both forward slashes and backslashes
func endsWithSeparator(path string) bool {
	return strings.HasSuffix(path, "/") || strings.HasSuffix(path, string(os.PathSeparator))
}
