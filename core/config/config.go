package config

type Config struct {
	SurfConfig   map[string]any `json:"surf" yaml:"surf"`
	ServerConfig map[string]any `json:"server" yaml:"server"`
}
