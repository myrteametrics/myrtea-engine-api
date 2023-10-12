package standalone

import (
	"errors"
	"go.uber.org/zap"
	"net/rpc"
)

// RPCServer This section concerns the plugin
type RPCServer struct {
	// This is the real implementation
	Impl StandaloneService
}

func (s *RPCServer) Run(args interface{}, resp *string) error {
	port, ok := args.(int)

	if !ok {
		return errors.New("couldn't cast port to int")
	}

	return s.Impl.Run(port)
}

// RPCPlugin This section concerns the engine
type RPCPlugin struct {
	client *rpc.Client
}

func (g *RPCPlugin) Run(port int) {
	err := g.client.Call("Plugin.Run", port, nil)
	if err != nil {
		zap.L().Error("Error executing RCPPlugin.Run", zap.Error(err))
	}
}
