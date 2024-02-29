package service

import (
	"errors"
	"github.com/myrteametrics/myrtea-sdk/v4/connector"
	"github.com/myrteametrics/myrtea-sdk/v4/expression"
	"testing"
)

type mockConnectorCaller struct {
	calls           map[string]int
	getRequestError error
}

func (m *mockConnectorCaller) getRequest(url, key string) ([]byte, error) {
	m.calls[url]++
	return nil, m.getRequestError
}

func (m *mockConnectorCaller) postRequest(url, key string) (int, error) {
	m.calls[url]++
	return 0, nil
}

func TestConnectorService_GetDefinition(t *testing.T) {
	service := &ConnectorService{
		Definition: Definition{
			Url: "http://localhost:0909",
			Key: "key",
		},
		ConnectorCaller: &mockConnectorCaller{},
	}

	def := service.GetDefinition()
	expression.AssertEqual(t, def.Url, "http://localhost:0909")
	expression.AssertEqual(t, def.Key, "key")
}

func TestConnectorService_Reload(t *testing.T) {
	mock := &mockConnectorCaller{calls: make(map[string]int)}
	service := &ConnectorService{
		Definition: Definition{
			Url: "http://localhost:0909",
			Key: "key",
		},
		ConnectorCaller: mock,
	}

	// no components test
	code, err := service.Reload("test")
	if err == nil || !errors.Is(err, connector.ReloaderComponentNotFoundErr) {
		t.Error("Expected error")
	}
	expression.AssertEqual(t, code, 429)

	// with components test
	service.Components = []string{"test"}

	code, err = service.Reload("test")
	if err != nil {
		t.Error("Unexpected error")
	}
	expression.AssertEqual(t, code, 0)
	expression.AssertEqual(t, mock.calls["http://localhost:0909/api/v1/reload"], 1)
}

func TestConnectorService_Restart(t *testing.T) {
	mock := &mockConnectorCaller{calls: make(map[string]int)}
	service := &ConnectorService{
		Definition: Definition{
			Url: "http://localhost:0909",
			Key: "key",
		},
		ConnectorCaller: mock,
	}

	_, _ = service.Restart()

	expression.AssertEqual(t, mock.calls["http://localhost:0909/api/v1/restart"], 1)
}

func TestConnectorService_GetStatus(t *testing.T) {
	mock := &mockConnectorCaller{calls: make(map[string]int)}
	service := &ConnectorService{
		Definition: Definition{
			Url: "http://localhost:0909",
			Key: "key",
		},
		ConnectorCaller: mock,
	}

	status := service.GetStatus()
	expression.AssertEqual(t, status.IsAlive, false)
	expression.AssertEqual(t, mock.calls["http://localhost:0909/api/v1/alive"], 1)
}

func TestConnectorService_GetStatus_WithError(t *testing.T) {
	mock := &mockConnectorCaller{calls: make(map[string]int), getRequestError: errors.New("error")}
	service := &ConnectorService{
		Definition: Definition{
			Url: "localhost",
			Key: "key",
		},
		ConnectorCaller: mock,
	}

	status := service.GetStatus()
	expression.AssertEqual(t, status.IsAlive, false)
}

func TestNewConnectorService(t *testing.T) {
	service := NewConnectorService(Definition{})
	expression.AssertNotEqual(t, service.Components, nil)

	service = NewConnectorService(Definition{Components: []string{"test"}})
	expression.AssertEqual(t, len(service.Components), 1)
	expression.AssertEqual(t, service.Components[0], "test")
}

func TestConnectorService_GetBaseUrl(t *testing.T) {
	service := NewConnectorService(Definition{
		Url: "blabla",
	})

	expression.AssertEqual(t, service.getBaseUrl(), "blabla/api/v1")
}
