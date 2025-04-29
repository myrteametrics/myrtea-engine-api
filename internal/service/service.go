package service

import (
	"github.com/google/uuid"
	"time"
)

type Definition struct {
	Id          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Url         string    `json:"-"`
	Key         string    `json:"-"`
	Type        string    `json:"type"`
	Components  []string  `json:"components"`
	LastRestart time.Time `json:"lastRestart"`
	LastReload  time.Time `json:"lastReload"`
}

type Status struct {
	IsAlive bool `json:"alive"`
}

type Service interface {
	GetStatus() Status
	Reload(component string) (int, error)
	GetDefinition() *Definition
	Restart() (int, error)
}

// HasComponent checks whether definition contains given component
func (d Definition) HasComponent(component string) bool {
	for _, c := range d.Components {
		if c == component {
			return true
		}
	}
	return false
}
