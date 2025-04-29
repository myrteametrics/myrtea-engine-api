package pluginutils

import "github.com/spf13/viper"

type PluginConfig struct {
	Name string
	Port int
}

// LoadPluginConfig Loads the plugin list from the TOML config file
func LoadPluginConfig() ([]PluginConfig, error) {
	var plugins []PluginConfig

	if !viper.IsSet("plugin") { // if no key set no plugins given, but no parse error
		return plugins, nil
	}

	err := viper.UnmarshalKey("plugin", &plugins)
	if err != nil {
		return nil, err
	}

	return plugins, nil
}
