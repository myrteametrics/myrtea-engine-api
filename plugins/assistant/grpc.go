package assistant

import (
	"github.com/hashicorp/go-plugin"
	"github.com/myrteametrics/myrtea-engine-api/v4/plugins/assistant/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type AssistantGRPCPlugin struct {
	// GRPCPlugin must still implement the Plugin interface
	plugin.Plugin
	// Concrete implementation, written in Go. This is only used for plugins that are written in Go.
	Impl Assistant
}

func (p *AssistantGRPCPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	proto.RegisterAssistantServer(s, &GRPCServer{Impl: p.Impl})
	return nil
}

func (p *AssistantGRPCPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCClient{client: proto.NewAssistantClient(c)}, nil
}

// GRPCClient is an implementation of Assistant that talks over RPC.
type GRPCClient struct {
	client proto.AssistantClient
}

func (m *GRPCClient) SentenceProcess(ti string, message string, tokens [][]string) ([]byte, []string, error) {

	previousContext := make([]*proto.TokenizedSentence, 0)
	for _, s := range tokens {
		previousContext = append(previousContext, &proto.TokenizedSentence{Word: s})
	}

	resp, err := m.client.SentenceProcess(context.Background(), &proto.NLPSentenceRequest{
		Time:            ti,
		Message:         message,
		PreviousContext: previousContext,
	})
	if err != nil {
		return nil, nil, err
	}

	return resp.Fact, resp.Tokens, nil
}

type GRPCServer struct {
	// This is the real implementation
	Impl Assistant
}

func (m *GRPCServer) SentenceProcess(ctx context.Context, req *proto.NLPSentenceRequest) (*proto.NLPSentenceResponse, error) {

	contextTokens := make([][]string, 0)
	for _, s := range req.PreviousContext {
		contextTokens = append(contextTokens, s.Word)
	}

	f, tokens, err := m.Impl.SentenceProcess(req.Time, req.Message, contextTokens)
	if err != nil {
		return nil, err
	}

	return &proto.NLPSentenceResponse{Fact: f, Tokens: tokens}, nil
}
