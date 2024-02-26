package service

import (
	"log"
	"net/http"
)

type ConnectorService struct {
	Definition
}

func (c *ConnectorService) Restart() error {
	return c.sendRequest()
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
	resp, err := http.Post("https://postman-echo.com/post", "application/json", nil)
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
