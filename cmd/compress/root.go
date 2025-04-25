// Package main implements the CLI for the video compression tool.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/yourorg/video-compressor/internal/presets"
	"github.com/yourorg/video-compressor/internal/queue"
)

var (
	// Command-line flags
	output     string // Output file or directory
	parallel   int    // Number of parallel compression jobs
	presetName string // Preset name to use
	configDir  string // Directory for preset YAML files
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		osExit(1)
	}
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "video-compress",
		Short: "Video Compressor: compress videos and manage presets",
	}

	// Global config directory for presets
	root.PersistentFlags().StringVarP(&configDir, "config", "c", "config/", "Directory for preset YAML files")

	// compress subcommand
	compressCmd := &cobra.Command{
		Use:   "compress [flags] <file1> [file2…]",
		Short: "Compress one or more videos",
		Args:  cobra.MinimumNArgs(1),
		RunE:  runCompress,
	}
	compressCmd.Flags().IntVarP(&parallel, "jobs", "j", 1, "Parallel compression jobs")
	compressCmd.Flags().StringVarP(&output, "output", "o", "", "Output file or directory")
	compressCmd.Flags().StringVarP(&presetName, "preset", "p", "default", "Preset name to use")
	// preset overrides
	compressCmd.Flags().String("video-codec", "", "Override video codec from preset")
	compressCmd.Flags().String("ffpreset", "", "Override ffmpeg preset from preset")
	compressCmd.Flags().Int("crf", -1, "Override CRF value from preset")

	root.AddCommand(compressCmd)
	root.AddCommand(newPresetsCmd())
	return root
}

func runCompress(cmd *cobra.Command, args []string) error {
	// Set config file path
	presets.ConfigFile = filepath.Join(configDir, "default.yaml")

	// 1. Load presets
	all, err := loadPresetsFunc()
	if err != nil {
		return fmt.Errorf("loading presets from %s: %w", presets.ConfigFile, err)
	}
	p, ok := all[presetName]
	if !ok {
		return fmt.Errorf("unknown preset %q; available: %v", presetName, presets.ListNames(all))
	}

	// 2. Apply overrides
	if vc := mustGetString(cmd, "video-codec"); vc != "" {
		p.VideoCodec = vc
	}
	if pf := mustGetString(cmd, "ffpreset"); pf != "" {
		p.Preset = pf
	}
	if crf := mustGetInt(cmd, "crf"); crf >= 0 {
		p.CRF = crf
	}

	// 3. Build ffmpeg args
	ffArgs := presets.BuildFFArgs(p)

	// 4. Initialize queue
	q := queue.New(parallel)

	// 5. Enqueue tasks
	for _, in := range args {
		dest := deriveOutput(in, output)
		q.Add(queue.Task{Source: in, Destination: dest, Args: ffArgs})
	}

	// 6. Run queue
	results := q.Run()

	// 7. Summarize
	successes, failures := 0, 0
	for _, r := range results {
		if r.Err != nil {
			fmt.Fprintf(os.Stderr, "Error compressing %s: %v\n", r.Task.Source, r.Err)
			failures++
		} else {
			cmd.Println("✓", r.Task.Source)
			successes++
		}
	}
	if failures > 0 {
		return fmt.Errorf("one or more files failed to compress")
	}
	cmd.Println("All done!")
	return nil
}

func newPresetsCmd() *cobra.Command {
	pcmd := &cobra.Command{
		Use:   "presets",
		Short: "Manage compression presets",
	}

	// Ensure using the correct config file
	pcmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		presets.ConfigFile = filepath.Join(configDir, "default.yaml")
	}

	// list
	pcmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all presets",
		RunE: func(cmd *cobra.Command, _ []string) error {
			all, err := loadPresetsFunc()
			if err != nil {
				return fmt.Errorf("loading presets from %s: %w", presets.ConfigFile, err)
			}
			for _, name := range presets.ListNames(all) {
				cmd.Println(name)
			}
			return nil
		},
	})

	// add
	add := &cobra.Command{
		Use:   "add <name>",
		Short: "Add or overwrite a preset",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("requires exactly 1 arg")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			p := presets.Preset{
				VideoCodec: mustGetString(cmd, "video-codec"),
				Preset:     mustGetString(cmd, "preset"),
				CRF:        mustGetInt(cmd, "crf"),
			}
			if err := savePresetFunc(args[0], p); err != nil {
				return fmt.Errorf("saving preset %s: %w", args[0], err)
			}
			return nil
		},
	}
	add.Flags().String("video-codec", "libx264", "ffmpeg video codec")
	add.Flags().String("preset", "medium", "ffmpeg preset")
	add.Flags().Int("crf", 23, "ffmpeg CRF value")
	pcmd.AddCommand(add)

	// remove
	rm := &cobra.Command{
		Use:   "remove <name>",
		Short: "Delete a named preset",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("requires exactly 1 arg")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := deletePresetFunc(args[0]); err != nil {
				return fmt.Errorf("deleting preset %s: %w", args[0], err)
			}
			return nil
		},
	}
	pcmd.AddCommand(rm)

	return pcmd
}

// mustGetString retrieves a string flag or panics
func mustGetString(cmd *cobra.Command, name string) string {
	s, err := cmd.Flags().GetString(name)
	if err != nil {
		panic(err)
	}
	return s
}

// mustGetInt retrieves an int flag or panics
func mustGetInt(cmd *cobra.Command, name string) int {
	i, err := cmd.Flags().GetInt(name)
	if err != nil {
		panic(err)
	}
	return i
}
