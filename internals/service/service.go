package service

import (
	"github.com/google/uuid"
	"time"
)

type Definition struct {
	Id         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	Url        string    `json:"-"`
	Port       int       `json:"-"`
	Key        string    `json:"-"`
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
