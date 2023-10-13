package standalone

import (
	"errors"
	"net/rpc"
)

// RPCServer This section concerns the plugin
type RPCServer struct {
	// This is the real implementation
	Impl StandaloneService
}

func (s *RPCServer) Run(args interface{}, resp *string) error {
	argsMap, ok := args.(map[string]interface{})

	if !ok {
		return errors.New("couldn't cast args to map[string]interface{}")
	}

	port, ok := argsMap["port"].(int)

	if !ok {
		return errors.New("couldn't cast port to int")
	}

	return s.Impl.Run(port)
}

// RPCPlugin This section concerns the engine
type RPCPlugin struct {
	client *rpc.Client
}

func (g *RPCPlugin) Run(port int) error {
	return g.client.Call("Plugin.Run", map[string]interface{}{
		"port": port,
	}, nil)
}
