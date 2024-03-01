package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/myrteametrics/myrtea-sdk/v4/connector"
	"go.uber.org/zap"
	"io"
	"net/http"
)

// ConnectorCaller is an interface to call a connector (mockable)
type ConnectorCaller interface {
	postRequest(url, key string) (int, error)
	getRequest(url, key string) ([]byte, error)
}

// httpConnectorCaller is a connector caller that uses HTTP
type httpConnectorCaller struct{}

type ConnectorService struct {
	Definition
	ConnectorCaller
}

// NewConnectorService creates a new connector service (with http connector caller)
func NewConnectorService(definition Definition) *ConnectorService {
	if definition.Components == nil {
		definition.Components = make([]string, 0)
	}
	return &ConnectorService{
		Definition:      definition,
		ConnectorCaller: &httpConnectorCaller{},
	}
}

// Restart restarts the connector
func (c *ConnectorService) Restart() (int, error) {
	return c.postRequest(fmt.Sprintf("%s/restart", c.getBaseUrl()), c.Key)
}

// GetStatus returns the status of the connector
func (c *ConnectorService) GetStatus() Status {
	status := Status{IsAlive: false}
	u := fmt.Sprintf("%s/isalive", c.getBaseUrl())

	body, err := c.getRequest(u, c.Key)
	if err != nil {
		zap.L().Error("GetConnectorStatus: could communicate with the connector", zap.Error(err), zap.String("url", u), zap.String("connectorName", c.Name))
		return Status{IsAlive: false}
	}

	err = json.Unmarshal(body, &status)
	if err != nil {
		zap.L().Error("GetConnectorStatus: could not unmarshall body", zap.Error(err), zap.String("connectorName", c.Name))
	}

	return status
}

// Reload reloads the connector
func (c *ConnectorService) Reload(component string) (int, error) {
	if !c.Definition.HasComponent(component) {
		return http.StatusTooManyRequests, connector.ReloaderComponentNotFoundErr
	}
	url := fmt.Sprintf("%s/reload/%s", c.getBaseUrl(), component)
	return c.postRequest(url, c.Key)
}

// GetDefinition returns the definition of the connector
func (c *ConnectorService) GetDefinition() *Definition {
	return &c.Definition
}

// getBaseUrl get connector base url
func (c *ConnectorService) getBaseUrl() string {
	return fmt.Sprintf("%s/api/v1", c.Url)
}

// postRequest calls the connector (in this case, an HTTP POST call)
func (h httpConnectorCaller) postRequest(url, key string) (int, error) {
	body := map[string]string{"key": key}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return 0, nil
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	return resp.StatusCode, nil
}

// callConnector calls the connector (in this case, an HTTP call)
func (h httpConnectorCaller) getRequest(url, key string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if key != "" {
		req.Header.Add("key", key)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
