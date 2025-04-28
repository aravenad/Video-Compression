// Package main implements the command-line interface for video compression.
package main

import (
	"os"

	"github.com/yourorg/video-compressor/internal/compressor"
	"github.com/yourorg/video-compressor/internal/presets"
	"github.com/yourorg/video-compressor/internal/queue"
)

// Hookable functions for test injection and dependency control:
// These variables allow tests to replace actual implementations with mocks.
var (
	loadPresetsFunc      = presets.LoadAll       // loads presets from config file
	compressFunc         = compressor.Compress   // performs video compression
	savePresetFunc       = presets.Save          // saves a preset to config
	deletePresetFunc     = presets.Delete        // removes a preset from config
	osExit               = os.Exit               // allows tests to prevent actual process termination
	getQueueCompressFunc = queue.GetCompressFunc // gets the queue compression function
	setQueueCompressFunc = queue.SetCompressFunc // sets the queue compression function
	mkdirAllFunc         = os.MkdirAll           // creates directories, mockable for tests
)
