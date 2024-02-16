package service

import (
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/models"
	plugin "github.com/myrteametrics/myrtea-engine-api/v5/plugins"
)

type PluginService struct {
	plugin plugin.Plugin
}

func (p *PluginService) GetStatus() models.ServiceStatus {
	return models.ServiceStatus{IsRunning: true}
}

func (p *PluginService) Reload(component string) error {
	return nil
}
