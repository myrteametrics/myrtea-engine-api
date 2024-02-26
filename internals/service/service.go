package service

import (
	"github.com/google/uuid"
	"time"
)

type Definition struct {
	Id         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	Hostname   string    `json:"-"`
	Port       int       `json:"-"`
	Type       string    `json:"type"`
	LastAction time.Time `json:"last-action"`
}

type Status struct {
	IsRunning bool `json:"running"`
}

type Service interface {
	GetStatus() Status
	Reload(component string) error
	GetDefinition() *Definition
	Restart() error
}
