package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/yourorg/video-compressor/internal/compressor"
)

var (
	input    string
	output   string
	preset   string
	parallel int
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "compress",
		Short: "Video compressor CLI",
		Long:  "A fast, single-binary Go CLI for cross-platform video compression.",
		RunE:  runCompress,
	}

	// Flags
	rootCmd.Flags().StringVarP(&input, "input", "i", "", "Input video file (required)")
	rootCmd.Flags().StringVarP(&output, "output", "o", "", "Output file or directory")
	rootCmd.Flags().StringVarP(&preset, "preset", "p", "default", "Compression preset name")
	rootCmd.Flags().IntVarP(&parallel, "jobs", "j", 1, "Number of parallel jobs")

	rootCmd.MarkFlagRequired("input")

	// Bind Viper to flags if needed
	viper.BindPFlag("preset", rootCmd.Flags().Lookup("preset"))

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runCompress(cmd *cobra.Command, args []string) error {
	// You can load config/default.yaml via Viper here, e.g.:
	viper.SetConfigName("default")
	viper.AddConfigPath("config")
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Stub: call into your compressor
	if err := compressor.Compress(input, output, viper.GetString("preset")); err != nil {
		return fmt.Errorf("compression failed: %w", err)
	}

	fmt.Println("Done!")
	return nil
}
