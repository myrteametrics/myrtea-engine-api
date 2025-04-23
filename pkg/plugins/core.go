package plugin

import (
	"fmt"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/plugins/baseline"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/plugins/pluginutils"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/plugins/standalone"
	"go.uber.org/zap"
	"net/http"
)

// MyrteaPlugin is a standard interface for any myrtea plugins
type MyrteaPlugin interface {
	ServicePort() int
	HandlerPrefix() string
	Handler() http.Handler
	Start() error
	Stop() error
	Running() bool
}

type Plugin struct {
	Config pluginutils.PluginConfig
	Plugin MyrteaPlugin
}

type Core struct {
	Plugins []Plugin
}

// RegisterPlugins Registers all plugins that were added into the TOML config file
func (c *Core) RegisterPlugins() {
	zap.L().Info("Registering plugins...")
	pluginConfigs, err := pluginutils.LoadPluginConfig()

	if err != nil {
		zap.L().Error("Couldn't parse plugin config", zap.Error(err))
		return
	}

	for _, config := range pluginConfigs {

		zap.L().Info(fmt.Sprintf("Registering plugin %s on port %d", config.Name, config.Port))

		if c.PluginExists(config.Name) {
			zap.L().Warn(fmt.Sprintf("Plugin %s already registered, skipping", config.Name))
			continue
		}

		plugin := Plugin{
			Config: config,
		}

		switch config.Name {
		case "baseline":
			if b := baseline.NewBaselinePlugin(config); b != nil {
				plugin.Plugin = b
			} else {
				continue
			}
			break
		default: // default is standalone plugins (no bi-directional communications needed)
			if s := standalone.NewPlugin(config); s != nil {
				plugin.Plugin = s
			} else {
				continue
			}
			break
		}

		c.Plugins = append(c.Plugins, plugin)
	}

	if len(c.Plugins) > 0 {
		zap.L().Info(fmt.Sprintf("Registered %d plugins !", len(c.Plugins)))
	} else {
		zap.L().Info("No plugins registered")
	}
}

// Start starts all plugins
func (c *Core) Start() {
	for _, plugin := range c.Plugins {
		err := plugin.Plugin.Start()
		if err != nil {
			zap.L().Error(fmt.Sprintf("Couldn't start plugin %s", plugin.Config.Name), zap.Error(err))
		}
	}
}

// Stop stops all plugins registered in core
func (c *Core) Stop() {
	for _, plugin := range c.Plugins {
		err := plugin.Plugin.Stop()
		if err != nil {
			zap.L().Error(fmt.Sprintf("Couldn't stop plugin %s", plugin.Config.Name), zap.Error(err))
		}
	}
}

// PluginExists checks if a plugin with the given name exists
func (c *Core) PluginExists(name string) (exists bool) {
	_, exists = c.GetPlugin(name)
	return
}

// GetPlugin returns a plugin by its name
func (c *Core) GetPlugin(name string) (MyrteaPlugin, bool) {
	for _, p := range c.Plugins {
		if p.Config.Name == name {
			return p.Plugin, true
		}
	}
	return nil, false
}
