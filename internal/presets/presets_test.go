package presets

import (
	"reflect"
	"testing"
)

func TestLoadAll(t *testing.T) {
	// Load presets from config/default.yaml
	all, err := LoadAll()
	if err != nil {
		t.Fatalf("LoadAll() error: %v", err)
	}

	p, ok := all["default"]
	if !ok {
		t.Fatalf("expected preset 'default' to exist")
	}

	// Verify fields match config/default.yaml
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
