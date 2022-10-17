package baseline

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/hashicorp/go-plugin"
	"github.com/myrteametrics/myrtea-engine-api/v5/plugins/pluginutils"
	"go.uber.org/zap"
)

var (
	_globalPluginMu sync.RWMutex
	_globalPlugin   *BaselinePlugin
)

// P is used to access the global plugin singleton
func P() (*BaselinePlugin, error) {
	_globalPluginMu.RLock()
	defer _globalPluginMu.RUnlock()

	plugin := _globalPlugin
	if plugin == nil {
		return nil, errors.New("no Baseline plugin found, feature is not available")
	}
	return plugin, nil
}

func Register(plugin *BaselinePlugin) func() {
	_globalPluginMu.Lock()
	defer _globalPluginMu.Unlock()

	prev := _globalPlugin
	_globalPlugin = plugin
	return func() { Register(prev) }
}

// Handshake is a common handshake that is shared by plugin and host.
var Handshake = plugin.HandshakeConfig{
	// This isn't required when using VersionedPlugins
	ProtocolVersion:  1,
	MagicCookieKey:   "BASIC_PLUGIN",
	MagicCookieValue: "hello",
}

var pluginServicePort int = 9082
var pluginName string = "baseline"

type BaselinePlugin struct {
	Name            string
	ClientConfig    *plugin.ClientConfig
	Client          *plugin.Client
	BaselineService BaselineService
}

func NewBaselinePlugin() *BaselinePlugin {
	pluginPath := fmt.Sprintf("plugin/myrtea-%s.plugin", pluginName)

	stat, err := os.Stat(pluginPath)
	if os.IsNotExist(err) || stat.IsDir() {
		return nil
	}

	cmd := exec.Command("sh", "-c", pluginPath)
	cmd.Env = os.Environ()
	// cmd.Env = append(cmd.Env, "MYRTEA_component_DEBUG_MODE=true")

	pluginMap := map[string]plugin.Plugin{
		pluginName: &BaselineGRPCPlugin{},
	}

	return &BaselinePlugin{
		Name: pluginName,
		ClientConfig: &plugin.ClientConfig{
			Logger:           pluginutils.ZapWrap(zap.L()),
			HandshakeConfig:  Handshake,
			Plugins:          pluginMap,
			Cmd:              cmd,
			AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		},
	}
}

func (p *BaselinePlugin) ServicePort() int {
	return pluginServicePort
}

func (p *BaselinePlugin) HandlerPrefix() string {
	return fmt.Sprintf("/%s", p.Name)
}

func (p *BaselinePlugin) Stop() error {
	p.Client.Kill()
	return nil
}

func (p *BaselinePlugin) Start() error {

	client := plugin.NewClient(p.ClientConfig)

	rpcClient, err := client.Client()
	if err != nil {
		zap.L().Error("Initialize rpc client", zap.String("module", pluginName), zap.Error(err))
		return err
	}

	raw, err := rpcClient.Dispense(pluginName)
	if err != nil {
		zap.L().Error("Dispense plugin", zap.String("module", pluginName), zap.Error(err))
		return err
	}

	p.BaselineService = raw.(BaselineService)
	p.Client = client

	Register(p)

	return nil
}

func (p *BaselinePlugin) Test() {
	result, err := p.BaselineService.GetBaselineValues(-1, 19, 4, 111, time.Now())
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(result)
}

func (p *BaselinePlugin) Handler() http.Handler {
	r := chi.NewRouter()

	// Add HTTP routes for every method exposed in the plugin interface GetBaselineValues

	return r
}
