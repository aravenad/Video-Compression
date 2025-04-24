package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/yourorg/video-compressor/internal/presets"
)

// global flags
var (
	output   string
	preset   string
	parallel int
)

// newRootCmd builds (but does not execute) the CLI command.
func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compress [flags] <file1> [file2] [...]",
		Short: "Video compressor CLI",
		Long:  "A fast, single-binary Go CLI for cross-platform video compression.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  runCompress,
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file or directory")
	cmd.Flags().StringVarP(&preset, "preset", "p", "default", "Compression preset name")
	cmd.Flags().IntVarP(&parallel, "jobs", "j", 1, "Number of parallel jobs")
	viper.BindPFlag("preset", cmd.Flags().Lookup("preset"))

	return cmd
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runCompress(cmd *cobra.Command, args []string) error {
	// 1. Load presets via the hookable func
	allPresets, err := loadPresetsFunc()
	if err != nil {
		return fmt.Errorf("loading presets: %w", err)
	}
	p, ok := allPresets[preset]
	if !ok {
		return fmt.Errorf("unknown preset %q; available: %v", preset, keys(allPresets))
	}

	// 2. Build ffmpeg args
	ffArgs := presets.BuildFFArgs(p)

	// 3. Concurrency
	sem := make(chan struct{}, parallel)
	errs := make(chan error, len(args))
	for _, infile := range args {
		sem <- struct{}{}
		go func(in string) {
			defer func() { <-sem }()
			out := deriveOutput(in, output)
			if err := compressFunc(in, out, ffArgs); err != nil {
				errs <- fmt.Errorf("%s: %w", in, err)
				return
			}
			// Report success into Cobra’s output writer
			cmd.Println("✓", in)
			errs <- nil
		}(infile)
	}

	// 4. Wait & report
	var failed bool
	for i := 0; i < len(args); i++ {
		if err := <-errs; err != nil {
			// Send errors to the command’s error writer
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			failed = true
		}
	}
	if failed {
		return fmt.Errorf("one or more files failed to compress")
	}

	// Final success message via Cobra
	cmd.Println("All done!")
	return nil
}

// keys as before...
func keys(m map[string]presets.Preset) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	return ks
}
