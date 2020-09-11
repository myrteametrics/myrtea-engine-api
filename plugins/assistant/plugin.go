package assistant

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/hashicorp/go-plugin"
	"github.com/myrteametrics/myrtea-sdk/v4/engine"
	"go.uber.org/zap"
)

// Handshake is a common handshake that is shared by plugin and host.
var Handshake = plugin.HandshakeConfig{
	// This isn't required when using VersionedPlugins
	ProtocolVersion:  1,
	MagicCookieKey:   "BASIC_PLUGIN",
	MagicCookieValue: "hello",
}

var pluginName string = "assistant"

type AssistantPlugin struct {
	Name         string
	ClientConfig *plugin.ClientConfig
	Client       *plugin.Client
	Assistant    Assistant
}

func NewAssistantPlugin() *AssistantPlugin {
	pluginPath := fmt.Sprintf("plugin/myrtea-%s.plugin", pluginName)

	stat, err := os.Stat(pluginPath)
	if os.IsNotExist(err) || stat.IsDir() {
		return nil
	}

	cmd := exec.Command("sh", "-c", pluginPath)
	cmd.Env = os.Environ()
	// cmd.Env = append(cmd.Env, "MYRTEA_ASSISTANT_DEBUG_MODE=true")

	pluginMap := map[string]plugin.Plugin{
		pluginName: &AssistantGRPCPlugin{},
	}

	return &AssistantPlugin{
		Name: pluginName,
		ClientConfig: &plugin.ClientConfig{
			HandshakeConfig:  Handshake,
			Plugins:          pluginMap,
			Cmd:              cmd,
			AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		},
	}
}

func (m *AssistantPlugin) HandlerPrefix() string {
	return fmt.Sprintf("/%s", m.Name)
}

func (m *AssistantPlugin) Stop() error {
	m.Client.Kill()
	return nil
}

func (m *AssistantPlugin) Start() error {

	client := plugin.NewClient(m.ClientConfig)

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

	m.Assistant = raw.(Assistant)
	m.Client = client

	return nil
}

func (m *AssistantPlugin) Test() {
	bFact, tokens, err := m.Assistant.SentenceProcess(
		"2020-10-03T12:30:00.000+02:00",
		"combien de colis pour client france",
		[][]string{{"combien", "colis", "for", "country", "espagne"}},
	)
	if err != nil {
		fmt.Println(err)
	}

	var f engine.Fact
	err = json.Unmarshal(bFact, &f)
	if err != nil {
		fmt.Println("error", err.Error())
		os.Exit(1)
	}

	zap.L().Info("process", zap.Any("fact", f), zap.Any("tokens", tokens))
}
