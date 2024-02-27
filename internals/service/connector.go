package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
)

type ConnectorService struct {
	Definition
}

func (c *ConnectorService) Restart() error {
	u, err := url.JoinPath(fmt.Sprintf("%s:%d", c.Url, c.Port), "api/v1", "restart")
	if err != nil {
		return err
	}

	body := map[string]string{"key": c.Key}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil
	}

	resp, err := http.Post(u, "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("invalid response status code '%d'", resp.StatusCode)
	}

	return nil
}

func (c *ConnectorService) GetStatus() Status {
	return Status{IsRunning: true}
}

func (c *ConnectorService) Reload(component string) error {
	return nil
}

func (c *ConnectorService) GetDefinition() *Definition {
	return &c.Definition
}

func (c *ConnectorService) sendRequest() error {
	u, err := url.JoinPath(fmt.Sprintf("%s:%d", c.Url, c.Port), "restart")
	if err != nil {
		return err
	}
	resp, err := http.Post(u, "application/json", nil)
	if err != nil {
		log.Fatalf("An Error Occured %v", err)
	}
	defer resp.Body.Close()
	//body, err := io.ReadAll(resp.Body)
	//if err != nil {
	//	log.Fatalln(err)
	//}
	return nil
}
