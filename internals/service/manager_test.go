package service

import (
	plugin "github.com/myrteametrics/myrtea-engine-api/v5/plugins"
	"github.com/myrteametrics/myrtea-engine-api/v5/plugins/pluginutils"
	"github.com/myrteametrics/myrtea-sdk/v4/expression"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"testing"
)

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Error("Manager is nil")
	}
}

func TestManager_Register(t *testing.T) {
	m := NewManager()
	s := &PluginService{}
	m.Register(s)
	if len(m.services) != 1 {
		t.Error("Service not registered")
	}
}

func TestManager_Get(t *testing.T) {
	m := NewManager()
	s := &PluginService{}
	m.Register(s)
	_, ok := m.Get(s.GetDefinition().Id)
	if !ok {
		t.Error("Service not found")
	}
}

func TestManager_GetAll(t *testing.T) {
	m := NewManager()
	s := &PluginService{}
	m.Register(s)
	services := m.GetAll()
	if len(services) != 1 {
		t.Error("Service not found")
	}
}

func TestManager_LoadPlugins(t *testing.T) {
	m := NewManager()
	err := m.LoadPlugins(nil)
	if err != nil {
		t.Error(err)
	}

	tmpDir, err := os.MkdirTemp("", "engine-api")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a file in the temporary directory including plugin configuration
	pluginConfig := `[[plugin]]
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

	// test with core & non-matching plugin
	core := &plugin.Core{}
	core.Plugins = append(core.Plugins, plugin.Plugin{
		Config: pluginutils.PluginConfig{
			Name: "test2",
			Port: 8080,
		},
	})
	err = m.LoadPlugins(core)

	expression.AssertEqual(t, err, nil, "No error expected")
	expression.AssertEqual(t, len(m.services), 0, "Plugin should not be registered")

	core = &plugin.Core{}
	core.Plugins = append(core.Plugins, plugin.Plugin{
		Config: pluginutils.PluginConfig{
			Name: "test",
			Port: 8080,
		},
	})

	err = m.LoadPlugins(core)
	expression.AssertEqual(t, err, nil, "No error expected")

	// Check if the plugin is registered

	expression.AssertEqual(t, len(m.services), 1, "Plugin not registered")

	def := m.GetAll()[0].GetDefinition()

	// Check if the plugin is registered with the correct name
	expression.AssertEqual(t, def.Name, "test", "Plugin not registered with the correct name")

	expression.AssertEqual(t, def.Port, 8080, "Plugin not registered with the correct port")

	// Check if the plugin is registered with the correct type
	expression.AssertEqual(t, def.Type, "plugin", "Plugin not registered with the correct type")

	// Emulate unmarshalling error
	viper.Set("plugin", "test")

	err = m.LoadPlugins(nil)
	expression.AssertNotEqual(t, err, nil, "Error expected")
}

func TestManager_LoadConnectors(t *testing.T) {
	m := NewManager()
	err := m.LoadConnectors()
	if err != nil {
		t.Error(err)
	}

	tmpDir, err := os.MkdirTemp("", "engine-api")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a file in the temporary directory including connector configuration
	connectorConfig := `[[connector]]
	name = "test"
	url = "http://localhost"
	port = 8080
	key = "testkey"`

	err = os.WriteFile(filepath.Join(tmpDir, "services.toml"), []byte(connectorConfig), 0644)
	expression.AssertEqual(t, err, nil, "No error expected")

	// load config from file
	viper.SetConfigName("services")
	viper.AddConfigPath(tmpDir)
	viper.SetConfigType("toml")
	err = viper.ReadInConfig()
	expression.AssertEqual(t, err, nil, "No error expected")

	err = m.LoadConnectors()
	expression.AssertEqual(t, err, nil, "No error expected")

	// Check if the connector is registered
	expression.AssertEqual(t, len(m.services), 1, "Connector not registered")

	def := m.GetAll()[0].GetDefinition()

	// Check if the connector is registered with the correct name
	expression.AssertEqual(t, def.Name, "test", "Connector not registered with the correct name")

	// Check if the connector is registered with the correct type
	expression.AssertEqual(t, def.Type, "connector", "Connector not registered with the correct type")

	// Check if the connector is registered with the correct key
	expression.AssertEqual(t, def.Key, "testkey", "Connector not registered with the corret key")

	// Emulate unmarshalling error
	viper.Set("connector", "test")

	err = m.LoadConnectors()
	expression.AssertNotEqual(t, err, nil, "Error expected")

}
