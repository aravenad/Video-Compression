// Tests for utility functions
package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDeriveOutput verifies the output path derivation logic for different scenarios:
// - No output specified
// - Output is a directory
// - Output has trailing slash
// - Output is an explicit filename
func TestDeriveOutput(t *testing.T) {
	// Prepare a real temp dir for the "existing directory" case
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		infile   string
		output   string
		wantEnds string // we'll just check suffix or exact match
	}{
		{
			name:     "no output → same dir with -compressed",
			infile:   filepath.Join("videos", "input.mp4"),
			output:   "",
			wantEnds: filepath.Join("videos", "input-compressed.mp4"),
		},
		{
			name:     "existing dir → inside that dir",
			infile:   "in.mov",
			output:   tmpDir,
			wantEnds: filepath.Join(tmpDir, "in.mov"),
		},
		{
			name:     "dir with trailing slash → inside",
			infile:   "a.avi",
			output:   tmpDir + string(os.PathSeparator),
			wantEnds: filepath.Join(tmpDir, "a.avi"),
		},
		{
			name:     "explicit filename",
			infile:   "one.webm",
			output:   "out.webm",
			wantEnds: "out.webm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deriveOutput(tt.infile, tt.output)
			if !filepath.IsAbs(tt.wantEnds) && got != tt.wantEnds {
				t.Errorf("deriveOutput(%q, %q) = %q; want %q", tt.infile, tt.output, got, tt.wantEnds)
			}
			// For the first case, wantEnds may be relative with directories; just ensure suffix
			if filepath.IsAbs(tt.wantEnds) {
				if got != tt.wantEnds {
					t.Errorf("deriveOutput(%q, %q) = %q; want %q", tt.infile, tt.output, got, tt.wantEnds)
				}
			} else {
				if !strings.HasSuffix(got, tt.wantEnds) {
					t.Errorf("deriveOutput(%q, %q) = %q; want suffix %q", tt.infile, tt.output, got, tt.wantEnds)
				}
			}
		})
	}
}
