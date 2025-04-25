// Package presets provides functionality for managing ffmpeg encoding presets
// stored in YAML configuration files.
package presets

import (
	"fmt"
	"os"
	"sort"

	"path/filepath"

	"gopkg.in/yaml.v3"
)

var (
	// ConfigFile is the path to the YAML file with presets.
	ConfigFile = "config/default.yaml"

	// yamlMarshal is a variable to allow mocking in tests
	yamlMarshal = yaml.Marshal
)

// Preset holds all the ffmpeg settings for one named preset.
// The Name field is not stored in YAML but populated from the map key.
type Preset struct {
	Name       string `yaml:"-"`
	VideoCodec string `mapstructure:"video_codec" yaml:"video_codec"`
	CRF        int    `mapstructure:"crf"           yaml:"crf"`
	Preset     string `mapstructure:"preset"        yaml:"preset"`
}

// LoadAll reads ConfigFile directly and unmarshals into a map of Presets.
// Returns error if the file cannot be read or parsed.
func LoadAll() (map[string]Preset, error) {
	// 1. Read the YAML file (may be test-overridden via ConfigFile)
	data, err := os.ReadFile(ConfigFile)
	if err != nil {
		return nil, fmt.Errorf("reading presets config: %w", err)
	}

	// 2. Unmarshal into an aux struct
	aux := struct {
		Presets map[string]Preset `yaml:"presets"`
	}{}
	if err := yaml.Unmarshal(data, &aux); err != nil {
		return nil, fmt.Errorf("parsing presets config: %w", err)
	}

	// 3. Wire up Name fields
	for name, p := range aux.Presets {
		p.Name = name
		aux.Presets[name] = p
	}

	return aux.Presets, nil
}

// BuildFFArgs turns a Preset into ffmpeg CLI args (minus input/output).
// Returns a slice of strings suitable for passing to exec.Command.
func BuildFFArgs(p Preset) []string {
	return []string{
		"-c:v", p.VideoCodec,
		"-preset", p.Preset,
		"-crf", fmt.Sprintf("%d", p.CRF),
	}
}

// Save writes or overwrites the preset named name to the config file.
// Creates the config file and directories if they don't exist.
func Save(name string, p Preset) error {
	// 1. Read existing file (or start fresh)
	cfg := struct {
		Presets map[string]Preset `yaml:"presets"`
	}{}
	data, err := os.ReadFile(ConfigFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("reading config: %w", err)
	}
	if len(data) > 0 {
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return fmt.Errorf("parsing config: %w", err)
		}
	}
	if cfg.Presets == nil {
		cfg.Presets = make(map[string]Preset)
	}

	// 2. Insert or overwrite
	p.Name = name
	cfg.Presets[name] = p

	// 3. Write back
	out, err := yamlMarshal(&cfg)
	if err != nil {
		return fmt.Errorf("serializing config: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(ConfigFile), 0o755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}
	if err := os.WriteFile(ConfigFile, out, 0o644); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}

// Delete removes the named preset from the config file.
// Returns error if the file doesn't exist or cannot be modified.
func Delete(name string) error {
	cfg := struct {
		Presets map[string]Preset `yaml:"presets"`
	}{}
	data, err := os.ReadFile(ConfigFile)
	if err != nil {
		return fmt.Errorf("reading config: %w", err)
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("parsing config: %w", err)
	}
	delete(cfg.Presets, name)

	out, err := yamlMarshal(&cfg)
	if err != nil {
		return fmt.Errorf("serializing config: %w", err)
	}
	if err := os.WriteFile(ConfigFile, out, 0o644); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	return nil
}

// ListNames returns the sorted list of preset names.
// This ensures consistent ordering for display purposes.
func ListNames(all map[string]Preset) []string {
	names := make([]string, 0, len(all))
	for n := range all {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}
