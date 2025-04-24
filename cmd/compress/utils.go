package main

import (
	"os"
	"path/filepath"
	"strings"
)

// deriveOutput returns the output file path for a given input file.
// Behavior:
// 1. If output == "": place alongside the input file, appending "-compressed" to the basename.
// 2. If output exists and is a directory, place the file inside it, keeping the same basename.
// 3. If output ends with a path separator, treat it as a directory.
// 4. Otherwise, treat output as the explicit file path.
func deriveOutput(infile, output string) string {
	// 1. No output provided: generate alongside infile
	if output == "" {
		dir := filepath.Dir(infile)
		ext := filepath.Ext(infile)
		name := strings.TrimSuffix(filepath.Base(infile), ext)
		return filepath.Join(dir, name+"-compressed"+ext)
	}

	// 2. If it's an existing directory, drop file inside
	if info, err := os.Stat(output); err == nil && info.IsDir() {
		return filepath.Join(output, filepath.Base(infile))
	}

	// 3. If it ends with a slash, treat as directory (even if not existing yet)
	if strings.HasSuffix(output, string(os.PathSeparator)) {
		return filepath.Join(output, filepath.Base(infile))
	}

	// 4. Otherwise, use exact output path
	return output
}
