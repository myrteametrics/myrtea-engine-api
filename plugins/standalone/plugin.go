package standalone

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"

	"github.com/go-chi/chi/v5"
	"github.com/hashicorp/go-plugin"
	"github.com/myrteametrics/myrtea-engine-api/v5/plugins/pluginutils"
	"go.uber.org/zap"
)

// Handshake is a common handshake that is shared by plugin and host.
var Handshake = plugin.HandshakeConfig{
	// This isn't required when using VersionedPlugins
	ProtocolVersion:  1,
	MagicCookieKey:   "BASIC_PLUGIN",
	MagicCookieValue: "hello",
}

type Plugin struct {
	Config       pluginutils.PluginConfig
	ClientConfig *plugin.ClientConfig
	Client       *plugin.Client
}

func NewPlugin(config pluginutils.PluginConfig) *Plugin {
	pluginPath := fmt.Sprintf("plugin/myrtea-%s.plugin", config.Name)

	stat, err := os.Stat(pluginPath)
	if os.IsNotExist(err) || stat.IsDir() {
		zap.L().Warn("Couldn't find plugin binaries", zap.String("pluginName", config.Name),
			zap.String("pluginPath", pluginPath))
		return nil
	}

	cmd := exec.Command("sh", "-c", pluginPath)
	cmd.Env = os.Environ()
	// cmd.Env = append(cmd.Env, "MYRTEA_component_DEBUG_MODE=true")

	pluginMap := map[string]plugin.Plugin{
		config.Name: &GRPCPlugin{},
	}

	return &Plugin{
		Config: config,
		ClientConfig: &plugin.ClientConfig{
			Logger:           pluginutils.ZapWrap(zap.L()),
			HandshakeConfig:  Handshake,
			Plugins:          pluginMap,
			Cmd:              cmd,
			AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		},
	}
}

func (p *Plugin) ServicePort() int {
	return p.Config.Port
}

func (p *Plugin) HandlerPrefix() string {
	return fmt.Sprintf("/%s", p.Config.Name)
}

func (p *Plugin) Stop() error {
	p.Client.Kill()
	return nil
}

func (p *Plugin) Start() error {

	client := plugin.NewClient(p.ClientConfig)

	rpcClient, err := client.Client()
	if err != nil {
		zap.L().Error("Initialize rpc client", zap.String("module", p.Config.Name), zap.Error(err))
		return err
	}

	raw, err := rpcClient.Dispense(p.Config.Name)
	if err != nil {
		zap.L().Error("Dispense plugin", zap.String("module", p.Config.Name), zap.Error(err))
		return err
	}

	_ = raw

	// p.BaselineService = raw.(BaselineService)
	p.Client = client

	return nil
}

func (p *Plugin) Handler() http.Handler {
	r := chi.NewRouter()

	// Add HTTP routes for every method exposed in the plugin interface GetBaselineValues

	return r
}
