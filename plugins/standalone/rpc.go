package standalone

import (
	"net/rpc"
)

// RPCServer This section concerns the plugin
type RPCServer struct {
	// This is the real implementation
	Impl StandaloneService
}

func (s *RPCServer) Run(port int, resp *interface{}) error {
	return s.Impl.Run(port)
}

// RPCPlugin This section concerns the engine
type RPCPlugin struct {
	client *rpc.Client
}

func (g *RPCPlugin) Run(port int) error {
	return g.client.Call("Plugin.Run", port, nil)
}
