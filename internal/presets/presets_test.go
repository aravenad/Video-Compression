// Package presets unit tests
package presets

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// TestLoadAll verifies that presets can be correctly loaded from a YAML file
func TestLoadAll(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "default.yaml")
	content := []byte(`presets:
  default:
    video_codec: libx264
    preset: medium
    crf: 23
`)
	if err := os.WriteFile(tmpFile, content, 0644); err != nil {
		t.Fatalf("Failed to create temp config: %v", err)
	}

	// Override the config path for this test
	oldConfigFile := ConfigFile
	ConfigFile = tmpFile
	defer func() { ConfigFile = oldConfigFile }()

	// Load presets from the temporary config file
	all, err := LoadAll()
	if err != nil {
		t.Fatalf("LoadAll() error: %v", err)
	}

	p, ok := all["default"]
	if !ok {
		t.Fatalf("expected preset 'default' to exist")
	}

	// Verify fields match
	if p.VideoCodec != "libx264" {
		t.Errorf("VideoCodec = %q; want %q", p.VideoCodec, "libx264")
	}
	if p.CRF != 23 {
		t.Errorf("CRF = %d; want %d", p.CRF, 23)
	}
	if p.Preset != "medium" {
		t.Errorf("Preset = %q; want %q", p.Preset, "medium")
	}
}

// TestLoadAll_FileNotFound verifies the error behavior when config file is missing
func TestLoadAll_FileNotFound(t *testing.T) {
	tmp := t.TempDir()
	ConfigFile = filepath.Join(tmp, "does-not-exist.yaml")

	_, err := LoadAll()
	if err == nil || !strings.Contains(err.Error(), "reading presets config") {
		t.Fatalf("expected a file-read error, got %v", err)
	}
}

// TestLoadAll_BadYAML verifies error handling with invalid YAML content
func TestLoadAll_BadYAML(t *testing.T) {
	tmp := t.TempDir()
	ConfigFile = filepath.Join(tmp, "bad.yaml")
	// write something that's not valid YAML
	os.WriteFile(ConfigFile, []byte("::: not yaml :::"), 0644)

	_, err := LoadAll()
	if err == nil || !strings.Contains(err.Error(), "parsing presets config") {
		t.Fatalf("expected a YAML parse error, got %v", err)
	}
}

// TestLoadAll_EmptyFile tests that an empty config file results in an empty map
func TestLoadAll_EmptyFile(t *testing.T) {
	tmp := t.TempDir()
	ConfigFile = filepath.Join(tmp, "empty.yaml")
	os.WriteFile(ConfigFile, []byte{}, 0644)
	all, err := LoadAll()
	if err != nil {
		t.Fatalf("empty file should yield empty map, got error: %v", err)
	}
	if len(all) != 0 {
		t.Errorf("expected no presets, got %v", all)
	}
}

// TestBuildFFArgs tests that ffmpeg arguments are correctly built from a preset
func TestBuildFFArgs(t *testing.T) {
	p := Preset{
		Name:       "default",
		VideoCodec: "libx264",
		CRF:        23,
		Preset:     "medium",
	}
	args := BuildFFArgs(p)
	want := []string{"-c:v", "libx264", "-preset", "medium", "-crf", "23"}

	if !reflect.DeepEqual(args, want) {
		t.Errorf("BuildFFArgs() = %v; want %v", args, want)
	}
}

// TestSaveAndDelete tests that presets can be saved to and deleted from config
func TestSaveAndDelete(t *testing.T) {
	// point ConfigFile at a temp file
	tmp := t.TempDir()
	ConfigFile = filepath.Join(tmp, "default.yaml")

	// 1) Save two presets
	a := Preset{VideoCodec: "x", Preset: "p", CRF: 10}
	b := Preset{VideoCodec: "y", Preset: "q", CRF: 20}
	if err := Save("a", a); err != nil {
		t.Fatalf("Save a failed: %v", err)
	}
	if err := Save("b", b); err != nil {
		t.Fatalf("Save b failed: %v", err)
	}

	// load back
	got, err := LoadAll()
	if err != nil {
		t.Fatalf("LoadAll after save: %v", err)
	}
	if _, ok := got["a"]; !ok {
		t.Errorf("got missing preset a")
	}
	if _, ok := got["b"]; !ok {
		t.Errorf("got missing preset b")
	}

	// 2) Delete one
	if err := Delete("a"); err != nil {
		t.Fatalf("Delete a failed: %v", err)
	}
	got2, err := LoadAll()
	if err != nil {
		t.Fatalf("LoadAll after delete: %v", err)
	}
	if _, ok := got2["a"]; ok {
		t.Errorf("preset a still present after delete")
	}
	if _, ok := got2["b"]; !ok {
		t.Errorf("preset b missing after delete")
	}
}

// TestListNames_Sorted verifies that preset names are returned in sorted order
func TestListNames_Sorted(t *testing.T) {
	m := map[string]Preset{"z": {}, "a": {}, "m": {}}
	got := ListNames(m)
	want := []string{"a", "m", "z"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ListNames() = %v; want %v", got, want)
	}
}

// TestSave_ReadConfigError tests error handling when the config file can't be read
func TestSave_ReadConfigError(t *testing.T) {
	// Create a temp file that we can't read
	tmpDir := t.TempDir()
	ConfigFile = filepath.Join(tmpDir, "readonly.yaml")

	// Write the file then make it readonly
	err := os.WriteFile(ConfigFile, []byte("bad content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// On Unix systems, simulate no read permissions
	// Note: This is platform-specific, so on Windows it might not actually block reading
	err = os.Chmod(ConfigFile, 0000)
	if err != nil {
		t.Skipf("Couldn't set permissions, skipping: %v", err)
	}

	// Make ConfigFile a directory to trigger a read error (works on most platforms)
	os.Remove(ConfigFile)
	err = os.Mkdir(ConfigFile, 0755)
	if err != nil {
		t.Skipf("Couldn't make ConfigFile a directory: %v", err)
	}

	err = Save("test", Preset{})
	if err == nil || !strings.Contains(err.Error(), "reading config") {
		t.Errorf("Expected 'reading config' error, got: %v", err)
	}
}

// TestSave_UnmarshalError tests error handling when config contains invalid YAML
func TestSave_UnmarshalError(t *testing.T) {
	tmpDir := t.TempDir()
	ConfigFile = filepath.Join(tmpDir, "corrupt.yaml")

	// Write invalid YAML content
	err := os.WriteFile(ConfigFile, []byte("{invalid: yaml: content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create corrupt config: %v", err)
	}

	err = Save("test", Preset{})
	if err == nil || !strings.Contains(err.Error(), "parsing config") {
		t.Errorf("Expected 'parsing config' error, got: %v", err)
	}
}

// TestDelete_ReadConfigError tests error handling when config can't be read for deletion
func TestDelete_ReadConfigError(t *testing.T) {
	tmpDir := t.TempDir()
	ConfigFile = filepath.Join(tmpDir, "nonexistent.yaml")

	err := Delete("test")
	if err == nil || !strings.Contains(err.Error(), "reading config") {
		t.Errorf("Expected 'reading config' error, got: %v", err)
	}
}

// TestDelete_UnmarshalError tests error handling when config has invalid YAML
func TestDelete_UnmarshalError(t *testing.T) {
	tmpDir := t.TempDir()
	ConfigFile = filepath.Join(tmpDir, "invalid.yaml")

	// Write invalid YAML content
	err := os.WriteFile(ConfigFile, []byte("not valid: : yaml"), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid config: %v", err)
	}

	err = Delete("test")
	if err == nil || !strings.Contains(err.Error(), "parsing config") {
		t.Errorf("Expected 'parsing config' error, got: %v", err)
	}
}

// TestSave_WriteError tests error handling when config can't be written
func TestSave_WriteError(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "readonly.yaml")

	// First create a valid config file
	err := os.WriteFile(configPath, []byte("presets: {}"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Make it read-only to cause a write error
	err = os.Chmod(configPath, 0400) // read-only
	if err != nil {
		t.Skipf("Couldn't set read-only permissions, skipping: %v", err)
	}

	// Now try to save to that read-only file
	oldConfig := ConfigFile
	ConfigFile = configPath
	defer func() { ConfigFile = oldConfig }()

	err = Save("test", Preset{})
	if err == nil {
		// On some systems (Windows), the OS might still allow writes despite permissions
		t.Skipf("Expected write error but got success - permission model may differ on this OS")
	} else if !strings.Contains(err.Error(), "writing config") {
		t.Errorf("Expected 'writing config' error, got: %v", err)
	}
}

// TestDelete_WriteError tests error handling when config can't be written after delete
func TestDelete_WriteError(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "readonly.yaml")

	// First create a valid YAML file with content
	content := []byte(`presets:
  test:
    video_codec: test
    preset: test
    crf: 10
`)
	err := os.WriteFile(configPath, content, 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Make it read-only to cause a write error
	err = os.Chmod(configPath, 0400) // read-only
	if err != nil {
		t.Skipf("Couldn't set read-only permissions, skipping: %v", err)
	}

	// Now try to delete from that read-only file
	oldConfig := ConfigFile
	ConfigFile = configPath
	defer func() { ConfigFile = oldConfig }()

	err = Delete("test")
	if err == nil {
		// On some systems (Windows), the OS might still allow writes despite permissions
		t.Skipf("Expected write error but got success - permission model may differ on this OS")
	} else if !strings.Contains(err.Error(), "writing config") {
		t.Errorf("Expected 'writing config' error, got: %v", err)
	}
}

// TestSave_MarshalError tests the serializing config error in Save
func TestSave_MarshalError(t *testing.T) {
	// Create a custom marshaler that always fails
	origMarshal := yamlMarshal
	yamlMarshal = func(in interface{}) ([]byte, error) {
		return nil, fmt.Errorf("mock marshal error")
	}
	defer func() { yamlMarshal = origMarshal }()

	tmpDir := t.TempDir()
	ConfigFile = filepath.Join(tmpDir, "config.yaml")

	err := Save("test", Preset{})
	if err == nil || !strings.Contains(err.Error(), "serializing config") {
		t.Errorf("Expected 'serializing config' error, got: %v", err)
	}
}

// TestDelete_MarshalError tests the serializing config error in Delete
func TestDelete_MarshalError(t *testing.T) {
	// Create a valid config file first
	tmpDir := t.TempDir()
	ConfigFile = filepath.Join(tmpDir, "config.yaml")
	content := []byte(`presets:
  test:
    video_codec: test
    preset: test
    crf: 10
`)
	err := os.WriteFile(ConfigFile, content, 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Replace Marshal function with one that fails
	origMarshal := yamlMarshal
	yamlMarshal = func(in interface{}) ([]byte, error) {
		return nil, fmt.Errorf("mock marshal error")
	}
	defer func() { yamlMarshal = origMarshal }()

	err = Delete("test")
	if err == nil || !strings.Contains(err.Error(), "serializing config") {
		t.Errorf("Expected 'serializing config' error, got: %v", err)
	}
}

// TestSave_MkdirError tests the directory creation error in Save
func TestSave_MkdirError(t *testing.T) {
	// Create a file that will prevent directory creation
	tmpDir := t.TempDir()
	blockingFile := filepath.Join(tmpDir, "blocking-file")
	err := os.WriteFile(blockingFile, []byte{}, 0644)
	if err != nil {
		t.Fatalf("Failed to create blocking file: %v", err)
	}

	// Try to create a directory with the same name as our file
	ConfigFile = filepath.Join(blockingFile, "config", "default.yaml")

	err = Save("test", Preset{})
	if err == nil || !strings.Contains(err.Error(), "creating config dir") {
		t.Errorf("Expected 'creating config dir' error, got: %v", err)
	}
}
