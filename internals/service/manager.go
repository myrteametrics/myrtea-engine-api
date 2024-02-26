package service

import (
	"github.com/google/uuid"
	plugin "github.com/myrteametrics/myrtea-engine-api/v5/plugins"
	"github.com/myrteametrics/myrtea-engine-api/v5/plugins/pluginutils"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Manager struct {
	services map[uuid.UUID]Service
}

func NewManager() *Manager {
	return &Manager{
		services: make(map[uuid.UUID]Service),
	}
}

// Register adds a service to the manager
func (m Manager) Register(service Service) {
	id := uuid.New()
	service.GetDefinition().Id = id
	m.services[id] = service
}

// Get returns a service by its name
func (m Manager) Get(id uuid.UUID) (Service, bool) {
	service, ok := m.services[id]
	return service, ok
}

// GetAll returns all services
func (m Manager) GetAll() (services []Service) {
	for _, s := range m.services {
		services = append(services, s)
	}
	return
}

// LoadConnectors loads the connectors from the config and register them
func (m Manager) LoadConnectors() error {
	var connectors []Definition

	if !viper.IsSet("connector") { // if no key set no plugins given, but no parse error
		return nil
	}

	err := viper.UnmarshalKey("connector", &connectors)
	if err != nil {
		return err
	}

	for _, c := range connectors {
		c.Type = "connector"
		m.Register(&ConnectorService{
			Definition: c,
		})
	}

	return nil
}

// LoadPlugins loads the plugins from the config and register them
func (m Manager) LoadPlugins(core *plugin.Core) error {
	plugins, err := pluginutils.LoadPluginConfig()
	if err != nil {
		return err
	}

	for _, p := range plugins {

		pluginService := &PluginService{
			Definition: Definition{
				Name:     p.Name,
				Hostname: "localhost",
				Port:     p.Port,
				Type:     "plugin",
			},
		}

		if core != nil {
			pl, ok := core.GetPlugin(p.Name)
			if !ok {
				zap.L().Warn("Trying to load a plugin that does not exist", zap.String("plugin", p.Name))
				continue
			}
			pluginService.MyrteaPlugin = pl
		}

		m.Register(pluginService)
	}

	return nil
}
