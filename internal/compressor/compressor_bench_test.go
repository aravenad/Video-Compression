package compressor

import (
	"path/filepath"
	"testing"

	"github.com/yourorg/video-compressor/internal/presets"
)

// BenchmarkCompressDefault measures the performance of the Compress function
// using the default preset on a sample video in testdata/.
func BenchmarkCompressDefault(b *testing.B) {
	infile := filepath.Join("testdata", "sample.mp4")
	outfile := filepath.Join("testdata", "out.mp4")

	// Build ffmpeg arguments for the default preset
	defaultPreset := presets.Preset{VideoCodec: "libx264", Preset: "medium", CRF: 23}
	args := presets.BuildFFArgs(defaultPreset)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := Compress(infile, outfile, args); err != nil {
			b.Fatal(err)
		}
	}
}
