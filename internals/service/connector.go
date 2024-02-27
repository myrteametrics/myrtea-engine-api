package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/myrteametrics/myrtea-sdk/v4/connector"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/url"
)

type ConnectorService struct {
	Definition
}

func (c *ConnectorService) Restart() (int, error) {
	hostname := fmt.Sprintf("%s:%d", c.Url, c.Port)
	u, err := url.JoinPath(hostname, "api", "v1", "restart")
	if err != nil {
		return 0, err
	}
	return c.callConnector(u)
}

func (c *ConnectorService) GetStatus() Status {
	status := Status{IsAlive: false}
	hostname := fmt.Sprintf("%s:%d", c.Url, c.Port)
	u, err := url.JoinPath(hostname, "api", "v1", "alive")
	if err != nil {
		zap.L().Error("GetConnectorStatus: Could not join path", zap.Error(err), zap.String("connectorName", c.Name))
		return status
	}

	resp, err := http.Get(u)
	if err != nil {
		zap.L().Error("GetConnectorStatus: could communicate with the connector", zap.Error(err), zap.String("url", u), zap.String("connectorName", c.Name))
		return Status{IsAlive: false}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		zap.L().Error("GetConnectorStatus: could not read body", zap.Error(err), zap.String("connectorName", c.Name))
		return Status{IsAlive: false}
	}

	err = json.Unmarshal(body, &status)
	if err != nil {
		zap.L().Error("GetConnectorStatus: could not unmarshall body", zap.Error(err), zap.String("connectorName", c.Name))
	}

	return status
}

func (c *ConnectorService) Reload(component string) (int, error) {
	if !c.Definition.HasComponent(component) {
		return http.StatusTooManyRequests, connector.ReloaderComponentNotFoundErr
	}
	hostname := fmt.Sprintf("%s:%d", c.Url, c.Port)
	u, err := url.JoinPath(hostname, "api", "v1", "reload", component)
	if err != nil {
		return 0, err
	}
	return c.callConnector(u)
}

func (c *ConnectorService) GetDefinition() *Definition {
	return &c.Definition
}

func (c *ConnectorService) callConnector(url string) (int, error) {
	body := map[string]string{"key": c.Key}
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
