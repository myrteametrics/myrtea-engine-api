package plugin

import (
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/plugins/pluginutils"
	"github.com/myrteametrics/myrtea-sdk/v5/expression"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"testing"
)

func TestCore_GetPlugin(t *testing.T) {
	c := Core{}
	_, ok := c.GetPlugin("test")
	if ok {
		t.Error("Plugin should not exist")
	}

	c.Plugins = append(c.Plugins, Plugin{
		Config: pluginutils.PluginConfig{
			Name: "test",
			Port: 8080,
		},
	})

	_, ok = c.GetPlugin("test")
	if !ok {
		t.Error("Plugin should exist")
	}
}

func TestCore_PluginExists(t *testing.T) {
	c := Core{}
	ok := c.PluginExists("test")
	if ok {
		t.Error("Plugin should not exist")
	}

	c.Plugins = append(c.Plugins, Plugin{
		Config: pluginutils.PluginConfig{
			Name: "test",
			Port: 8080,
		},
	})

	ok = c.PluginExists("test")
	if !ok {
		t.Error("Plugin should exist")
	}
}

func TestCore_RegisterPlugins(t *testing.T) {
	c := Core{}
	tmpDir, err := os.MkdirTemp("", "engine-api")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a file in the temporary directory including plugin configuration
	pluginConfig := `[[plugin]]
	name = "test"
	port = 8080

	[[plugin]]
	name = "test"
	port = 8080`

	err = os.WriteFile(filepath.Join(tmpDir, "services.toml"), []byte(pluginConfig), 0644)
	expression.AssertEqual(t, err, nil, "No error expected")

	// load config from file
	viper.SetConfigName("services")
	viper.AddConfigPath(tmpDir)
	viper.SetConfigType("toml")
	err = viper.ReadInConfig()
	expression.AssertEqual(t, err, nil, "No error expected")

	c.RegisterPlugins()
	expression.AssertEqual(t, len(c.Plugins), 0, "No plugins should be registered since plugin does not exists")
	viper.Set("plugin", "test")
	c = Core{}
	c.RegisterPlugins()
	expression.AssertEqual(t, len(c.Plugins), 0, "No plugin should be registered")

}

func TestCore_Start(t *testing.T) {
	c := Core{}
	c.Start()
}

func TestCore_Stop(t *testing.T) {
	c := Core{}
	c.Stop()
}
