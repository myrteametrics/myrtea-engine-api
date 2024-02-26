package service

import (
	plugin "github.com/myrteametrics/myrtea-engine-api/v5/plugins"
)

type PluginService struct {
	Definition
	plugin.MyrteaPlugin
}

func (p *PluginService) GetStatus() Status {
	return Status{IsRunning: true}
}

// Reload reloads the service
func (p *PluginService) Reload(component string) error {
	// TODO: implement reload
	return nil
}

// GetDefinition returns the definition of the service
func (p *PluginService) GetDefinition() *Definition {
	return &p.Definition
}

// Restart restarts the service
func (p *PluginService) Restart() error {
	err := p.Stop()
	if err != nil {
		return err
	}

	err = p.Start()
	if err != nil {
		return err
	}
	return nil
}
