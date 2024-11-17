package config

type Config struct {
	CoreConfig   map[string]any `json:"core" yaml:"core"`
	ServerConfig map[string]any `json:"server" yaml:"server"`
}
