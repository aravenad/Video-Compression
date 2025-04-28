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
			// Add a specific test for trailing forward slash (Case 3)
			name:     "dir with forward slash → inside",
			infile:   "b.mp4",
			output:   "output/",
			wantEnds: filepath.Join("output", "b.mp4"),
		},
		{
			// Test multiple trailing separators are handled properly
			name:     "dir with multiple trailing slashes → inside",
			infile:   "c.mp4",
			output:   "multiple//",
			wantEnds: filepath.Join("multiple", "c.mp4"),
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
			expectedPath := filepath.Clean(tt.wantEnds) // Ensure consistent path format

			if filepath.IsAbs(expectedPath) {
				if got != expectedPath {
					t.Errorf("deriveOutput(%q, %q) = %q; want %q", tt.infile, tt.output, got, expectedPath)
				}
			} else {
				// For relative paths, just check the base filename is correct
				gotBase := filepath.Base(got)
				expectedBase := filepath.Base(expectedPath)

				if gotBase != expectedBase {
					t.Errorf("deriveOutput(%q, %q) base filename = %q; want %q", tt.infile, tt.output, gotBase, expectedBase)
				}

				// For case 3 & 4, also verify the directory part is correct
				if strings.Contains(tt.name, "forward slash") || strings.Contains(tt.name, "multiple trailing") {
					gotDir := filepath.Dir(got)
					expectedDir := filepath.Dir(expectedPath)
					if gotDir != expectedDir {
						t.Errorf("deriveOutput(%q, %q) directory = %q; want %q", tt.infile, tt.output, gotDir, expectedDir)
					}
				}
			}
		})
	}
}

// TestDeriveOutput_TrimTrailingSeparators explicitly tests the trimming of trailing separators
func TestDeriveOutput_TrimTrailingSeparators(t *testing.T) {
	// Test cases for trailing separators
	tests := []struct {
		name   string
		infile string
		output string
	}{
		{
			name:   "forward slash",
			infile: "test.mp4",
			output: "dir/",
		},
		{
			name:   "backslash",
			infile: "test.mp4",
			output: `dir\`,
		},
		{
			name:   "multiple slashes",
			infile: "test.mp4",
			output: "dir///",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deriveOutput(tt.infile, tt.output)

			// Instead of checking the exact path (which can vary by platform),
			// check that we got "dir/test.mp4" structure
			gotDir := filepath.Dir(got)
			expectedDir := filepath.Clean("dir")

			if gotDir != expectedDir {
				t.Errorf("deriveOutput(%q, %q) directory = %q; want %q",
					tt.infile, tt.output, gotDir, expectedDir)
			}

			gotFile := filepath.Base(got)
			if gotFile != tt.infile {
				t.Errorf("deriveOutput(%q, %q) filename = %q; want %q",
					tt.infile, tt.output, gotFile, tt.infile)
			}
		})
	}
}
