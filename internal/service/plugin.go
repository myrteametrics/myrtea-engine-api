package service

import (
	plugin "github.com/myrteametrics/myrtea-engine-api/v5/pkg/plugins"
)

type PluginService struct {
	Definition
	plugin.MyrteaPlugin
}

func (p *PluginService) GetStatus() Status {
	return Status{IsAlive: p.MyrteaPlugin.Running()}
}

// Reload reloads the service
func (p *PluginService) Reload(component string) (int, error) {
	// TODO: implement reload
	return 0, nil
}

// GetDefinition returns the definition of the service
func (p *PluginService) GetDefinition() *Definition {
	return &p.Definition
}

// Restart restarts the service
func (p *PluginService) Restart() (int, error) {
	err := p.Stop()
	if err != nil {
		return 0, err
	}

	err = p.Start()
	if err != nil {
		return 0, err
	}
	return 200, nil
}
