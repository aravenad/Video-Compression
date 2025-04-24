package presets

import (
	"fmt"

	"github.com/spf13/viper"
)

// Preset holds all the ffmpeg settings for one named preset.
type Preset struct {
	Name       string
	VideoCodec string `mapstructure:"video_codec"`
	CRF        int    `mapstructure:"crf"`
	Preset     string `mapstructure:"preset"` // e.g. “medium”, “fast”
}

// LoadAll reads config/default.yaml into Viper and unmarshals into a map of Presets.
// It searches for the config file in "config/", "../config/", and "../../config/" to support tests.
func LoadAll() (map[string]Preset, error) {
	viper.SetConfigName("default")
	// Search project config directory when running at root
	viper.AddConfigPath("config")
	// Support tests running from internal/presets
	viper.AddConfigPath("../config")
	// Support tests running from internal/presets subdir on different CWD
	viper.AddConfigPath("../../config")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("reading presets config: %w", err)
	}

	raw := viper.GetStringMap("presets")
	result := make(map[string]Preset, len(raw))

	for name := range raw {
		sub := viper.Sub("presets." + name)
		if sub == nil {
			continue
		}
		var p Preset
		if err := sub.Unmarshal(&p); err != nil {
			return nil, fmt.Errorf("unmarshal preset %q: %w", name, err)
		}
		p.Name = name
		result[name] = p
	}
	return result, nil
}

// BuildFFArgs turns a Preset into ffmpeg CLI args (minus input/output).
func BuildFFArgs(p Preset) []string {
	return []string{
		"-c:v", p.VideoCodec,
		"-preset", p.Preset,
		"-crf", fmt.Sprintf("%d", p.CRF),
	}
}
