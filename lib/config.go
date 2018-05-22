package lib

import "github.com/BurntSushi/toml"

type Config struct {
	Main    MainConfig              `toml:"main"`
	Ledisdb LedisdbConfig           `toml:"ledisdb"`
	Plugin  map[string]PluginConfig `toml:"plugin"`
}

type MainConfig struct {
	Schedule string `toml:"schedule"`
	Debug    bool   `toml:"debug"`
}

type LedisdbConfig struct {
	Addr     string `toml:"addr"`
	Password string `toml:"password"`
	DB       int    `toml:"db"`
}

type PluginConfig struct {
	Path    string `toml:"path"`
	ListKey string `toml:"list_key"`
}

// DecodeConfigToml ...
func DecodeConfigToml(tomlfile string) (Config, error) {
	var config Config
	_, err := toml.DecodeFile(tomlfile, &config)
	if err != nil {
		return config, err
	}
	return config, nil
}
