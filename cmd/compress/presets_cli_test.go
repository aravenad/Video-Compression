// Tests for the presets CLI commands
package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/yourorg/video-compressor/internal/presets"
)

// TestCLI_PresetsList verifies the presets list command outputs preset names
func TestCLI_PresetsList(t *testing.T) {
	// stub out loadPresetsFunc
	origLoad := loadPresetsFunc
	defer func() { loadPresetsFunc = origLoad }()
	loadPresetsFunc = func() (map[string]presets.Preset, error) {
		return map[string]presets.Preset{"foo": {}, "bar": {}}, nil
	}

	cmd := newRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"presets", "list"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("presets list failed: %v", err)
	}
	got := strings.Fields(buf.String())
	sort.Strings(got)
	want := []string{"bar", "foo"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("presets list = %v; want %v", got, want)
	}
}

// TestCLI_PresetsList_LoadError verifies error handling when presets can't be loaded
func TestCLI_PresetsList_LoadError(t *testing.T) {
	orig := loadPresetsFunc
	defer func() { loadPresetsFunc = orig }()
	loadPresetsFunc = func() (map[string]presets.Preset, error) {
		return nil, fmt.Errorf("boom")
	}

	cmd := newRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"presets", "list"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("expected load error, got %v", err)
	}
}

// TestCLI_PresetsAddAndRemove verifies the add and remove commands work correctly
func TestCLI_PresetsAddAndRemove(t *testing.T) {
	// stub savePresetFunc and deletePresetFunc
	origSave := savePresetFunc
	origDel := deletePresetFunc
	defer func() {
		savePresetFunc = origSave
		deletePresetFunc = origDel
	}()

	var saved, deleted []string
	savePresetFunc = func(name string, p presets.Preset) error {
		saved = append(saved, name)
		// verify that flags are parsed into the preset struct
		if p.VideoCodec != "h264" || p.Preset != "fast" || p.CRF != 30 {
			t.Errorf("unexpected preset values: %+v", p)
		}
		return nil
	}
	deletePresetFunc = func(name string) error {
		deleted = append(deleted, name)
		return nil
	}

	cmd := newRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	// test add
	cmd.SetArgs([]string{
		"presets", "add", "new-preset",
		"--video-codec", "h264",
		"--preset", "fast",
		"--crf", "30",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("presets add failed: %v", err)
	}
	if !reflect.DeepEqual(saved, []string{"new-preset"}) {
		t.Errorf("saved = %v; want [new-preset]", saved)
	}

	// test remove
	buf.Reset()
	cmd.SetArgs([]string{"presets", "remove", "new-preset"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("presets remove failed: %v", err)
	}
	if !reflect.DeepEqual(deleted, []string{"new-preset"}) {
		t.Errorf("deleted = %v; want [new-preset]", deleted)
	}
}

// TestCLI_PresetsAdd_Failure verifies error handling when preset can't be saved
func TestCLI_PresetsAdd_Failure(t *testing.T) {
	origSave := savePresetFunc
	defer func() { savePresetFunc = origSave }()

	savePresetFunc = func(name string, p presets.Preset) error {
		return errors.New("disk full")
	}

	cmd := newRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"presets", "add", "x"})

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "disk full") {
		t.Fatalf("expected disk full error, got %v", err)
	}
}

// TestCLI_PresetsRemove_Failure verifies error handling when preset can't be removed
func TestCLI_PresetsRemove_Failure(t *testing.T) {
	origDel := deletePresetFunc
	defer func() { deletePresetFunc = origDel }()

	deletePresetFunc = func(name string) error {
		return errors.New("perm denied")
	}

	cmd := newRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"presets", "remove", "x"})

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "perm denied") {
		t.Fatalf("expected perm denied error, got %v", err)
	}
}

// TestCLI_PresetsUsage verifies the commands validate their arguments correctly
func TestCLI_PresetsUsage(t *testing.T) {
	cmd := newRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	// add without name
	cmd.SetArgs([]string{"presets", "add"})
	if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "requires exactly 1 arg") {
		t.Errorf("got %v; want argument count error", err)
	}
	// remove without name
	buf.Reset()
	cmd.SetArgs([]string{"presets", "remove"})
	if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "requires exactly 1 arg") {
		t.Errorf("got %v; want argument count error", err)
	}
}

// TestCLI_Compress_NoArgs verifies error handling when no files are provided
func TestCLI_Compress_NoArgs(t *testing.T) {
	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"compress"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected an error when no files are passed")
	}
	if !strings.Contains(err.Error(), "requires at least 1 arg") {
		t.Errorf("got %v; want \"requires at least 1 arg\"", err)
	}
}

// TestCLI_Compress_WithJobsAndOutput verifies options like parallel jobs and output directory
func TestCLI_Compress_WithJobsAndOutput(t *testing.T) {
	// Stub both the main package and queue package compressFunc variables
	origLoad := loadPresetsFunc
	origCompress := compressFunc

	// Get the compressFunc from the queue package
	origQueueCompress := getQueueCompressFunc()

	defer func() {
		loadPresetsFunc = origLoad
		compressFunc = origCompress
		setQueueCompressFunc(origQueueCompress)
	}()

	loadPresetsFunc = func() (map[string]presets.Preset, error) {
		return map[string]presets.Preset{"default": {}}, nil
	}

	// Create a mock compress function that just validates paths and returns success
	mockCompressFunc := func(in, out string, args []string) error {
		// With "--output", "outdir/", we should get outdir/filename.mp4
		if !strings.HasPrefix(out, "outdir"+string(os.PathSeparator)) {
			return fmt.Errorf("unexpected output path: %s", out)
		}
		return nil
	}

	// Set the mock in both places
	compressFunc = mockCompressFunc
	setQueueCompressFunc(mockCompressFunc)

	cmd := newRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{
		"compress",
		"--jobs", "3",
		"--output", "outdir/",
		"foo.mp4", "bar.mp4",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "✓ foo.mp4") || !strings.Contains(out, "✓ bar.mp4") {
		t.Errorf("unexpected output:\n%s", out)
	}
}
