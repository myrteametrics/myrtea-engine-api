package standalone

import (
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/hashicorp/go-plugin"
	pluginutils2 "github.com/myrteametrics/myrtea-engine-api/v5/pkg/plugins/pluginutils"
	"go.uber.org/zap"
	"net/http"
	"net/rpc"
	"os"
	"os/exec"
	"runtime"
)

// Handshake is a common handshake that is shared by plugin and host.
var Handshake = plugin.HandshakeConfig{
	// This isn't required when using VersionedPlugins
	ProtocolVersion:  1,
	MagicCookieKey:   "BASIC_PLUGIN",
	MagicCookieValue: "hello",
}

type Plugin struct {
	Config       pluginutils2.PluginConfig
	ClientConfig *plugin.ClientConfig
	client       *plugin.Client
	Impl         StandaloneService
}

func NewPlugin(config pluginutils2.PluginConfig) *Plugin {
	pluginPath := fmt.Sprintf("plugin/myrtea-%s.plugin", config.Name)

	stat, err := os.Stat(pluginPath)
	if os.IsNotExist(err) || stat.IsDir() {
		zap.L().Warn("Couldn't find plugin binaries", zap.String("pluginName", config.Name),
			zap.String("pluginPath", pluginPath))
		return nil
	}

	return &Plugin{Config: config}
}

func (p *Plugin) init() {
	pluginPath := fmt.Sprintf("plugin/myrtea-%s.plugin", p.Config.Name)

	// since plugin can be restarted, implement here the cmd instance creation every other start
	var cmd *exec.Cmd
	if runtime.GOOS != "windows" {
		cmd = exec.Command("sh", "-c", pluginPath)
	} else {
		cmd = exec.Command(pluginPath)
	}
	cmd.Env = os.Environ()
	// cmd.Env = append(cmd.Env, "MYRTEA_component_DEBUG_MODE=true")

	p.ClientConfig = &plugin.ClientConfig{
		Logger:          pluginutils2.ZapWrap(zap.L()),
		HandshakeConfig: Handshake,
		Plugins: map[string]plugin.Plugin{
			p.Config.Name: &Plugin{},
		},
		Cmd:              cmd,
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolNetRPC},
	}
}

func (p *Plugin) ServicePort() int {
	return p.Config.Port
}

func (p *Plugin) HandlerPrefix() string {
	return fmt.Sprintf("/%s", p.Config.Name)
}

func (p *Plugin) Stop() error {
	if p.client == nil || p.client.Exited() {
		return nil
	}
	p.client.Kill()
	return nil
}

func (p *Plugin) Start() error {
	if p.Running() {
		return errors.New("plugin is already running")
	}

	// first init plugin
	p.init()

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

	p.client = client

	service, ok := raw.(StandaloneService)
	if !ok {
		zap.L().Error("Cast dispense plugin", zap.String("module", p.Config.Name))
		client.Kill()
		return fmt.Errorf("could'nt cast dispense to StandaloneService")
	}

	p.Impl = service

	// Run service on given port
	errStr := p.Impl.Run(p.Config.Port)
	if errStr != "" {
		zap.L().Error("Error executing RCPPlugin.Run", zap.String("error", errStr))
		client.Kill()
		return err
	}

	return nil
}

func (p *Plugin) Running() bool {
	return p.client != nil && !p.client.Exited()
}

func (p *Plugin) Handler() http.Handler {
	r := chi.NewRouter()
	return r
}

// RPC server & client implementation for Plugin interface
func (p *Plugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &RPCServer{Impl: p.Impl}, nil
}

func (Plugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &RPCPlugin{client: c}, nil
}
