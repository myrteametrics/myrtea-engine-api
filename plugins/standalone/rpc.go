package standalone

import (
	"net/rpc"
)

// RPCServer This section concerns the plugin
type RPCServer struct {
	// This is the real implementation
	Impl StandaloneService
}

func (s *RPCServer) Run(port int, resp *string) error {
	*resp = s.Impl.Run(port)
	return nil
}

// RPCPlugin This section concerns the engine
type RPCPlugin struct {
	client *rpc.Client
}

func (g *RPCPlugin) Run(port int) string {
	var result string
	err := g.client.Call("Plugin.Run", port, &result)
	if err != nil {
		result = err.Error()
	}
	return result
}
