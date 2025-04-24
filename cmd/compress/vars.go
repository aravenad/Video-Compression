package main

import (
	"github.com/yourorg/video-compressor/internal/compressor"
	"github.com/yourorg/video-compressor/internal/presets"
)

// Hookable functions for test injection:
var (
	// loadPresetsFunc is a function that loads all presets from the config file.
	// It can be replaced in tests to return a different set of presets.
	loadPresetsFunc = presets.LoadAll
	// compressFunc is a function that compresses a video file using ffmpeg.
	// It can be replaced in tests to simulate compression without actually running ffmpeg.
	compressFunc = compressor.Compress
)
