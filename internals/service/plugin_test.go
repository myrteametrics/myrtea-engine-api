package service

import (
	"errors"
	"github.com/google/uuid"
	"github.com/myrteametrics/myrtea-sdk/v4/expression"
	"net/http"
	"testing"
)

type MyrteaPluginMock struct {
	calls     map[string]int
	stopError error
}

func (p *MyrteaPluginMock) ServicePort() int {
	//TODO implement me
	panic("implement me")
}

func (p *MyrteaPluginMock) HandlerPrefix() string {
	//TODO implement me
	panic("implement me")
}

func (p *MyrteaPluginMock) Handler() http.Handler {
	//TODO implement me
	panic("implement me")
}

func (p *MyrteaPluginMock) Running() bool {
	p.calls["Running"]++
	return true
}

func (p *MyrteaPluginMock) Start() error {
	p.calls["Start"]++
	return nil
}

func (p *MyrteaPluginMock) Stop() error {
	p.calls["Stop"]++
	return p.stopError
}

func TestPluginService_GetDefinition(t *testing.T) {
	uid := uuid.New()
	service := PluginService{
		Definition: Definition{
			Id: uid,
		},
	}
	definition := service.GetDefinition()
	expression.AssertEqual(t, definition.Id, uid)
}

func TestPluginService_GetStatus(t *testing.T) {
	mock := &MyrteaPluginMock{calls: make(map[string]int)}
	service := PluginService{
		MyrteaPlugin: mock,
	}
	status := service.GetStatus()
	expression.AssertEqual(t, status.IsAlive, true)
	expression.AssertEqual(t, mock.calls["Running"], 1)
}

func TestPluginService_Reload(t *testing.T) {
	//TODO: implement reload
	service := PluginService{}
	status, err := service.Reload("component")
	expression.AssertEqual(t, err, nil)
	expression.AssertEqual(t, status, 0)
}

func TestPluginService_Restart(t *testing.T) {
	mock := &MyrteaPluginMock{calls: make(map[string]int)}
	service := PluginService{
		MyrteaPlugin: mock,
	}
	status, err := service.Restart()
	expression.AssertEqual(t, err, nil)
	expression.AssertEqual(t, status, 0)
	expression.AssertEqual(t, mock.calls["Start"], 1)
	expression.AssertEqual(t, mock.calls["Stop"], 1)
}

func TestPluginService_Restart_Error(t *testing.T) {
	stopError := errors.New("error")
	mock := &MyrteaPluginMock{calls: make(map[string]int), stopError: stopError}
	service := PluginService{
		MyrteaPlugin: mock,
	}
	status, err := service.Restart()
	expression.AssertEqual(t, err, stopError)
	expression.AssertEqual(t, status, 0)
	expression.AssertEqual(t, mock.calls["Start"], 0)
	expression.AssertEqual(t, mock.calls["Stop"], 1)
}
