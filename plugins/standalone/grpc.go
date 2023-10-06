package standalone

import "github.com/hashicorp/go-plugin"

type GRPCPlugin struct {
	// GRPCPlugin must still implement the Plugin interface
	plugin.Plugin
}

//func (p *GRPCPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
//	return &GRPCPlugin{}, nil
//}
//
//func (GRPCPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
//	return &Plugin{ClientConfig: c}, nil
//}
