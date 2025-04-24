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
	// you can add AudioCodec, Bitrate, etc.
}

// LoadAll reads config/default.yaml (or other config files) into Viper
// and unmarshals into a map of Presets.
func LoadAll() (map[string]Preset, error) {
	viper.SetConfigName("default")
	viper.AddConfigPath("config")
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("reading presets config: %w", err)
	}

	raw := viper.GetStringMap("presets")
	result := make(map[string]Preset, len(raw))

	for name := range raw {
		var p Preset
		if err := viper.Sub("presets." + name).Unmarshal(&p); err != nil {
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
		// add audio flags or other filters here
	}
}
