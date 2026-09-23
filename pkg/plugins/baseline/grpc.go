package baseline

import (
	"time"

	proto2 "github.com/myrteametrics/myrtea-engine-api/v5/pkg/plugins/baseline/proto"

	"github.com/hashicorp/go-plugin"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const timeLayout = "2006-01-02T15:04:05.000Z07:00"

type BaselineGRPCPlugin struct {
	// GRPCPlugin must still implement the Plugin interface
	plugin.Plugin
	// Concrete implementation, written in Go. This is only used for plugins that are written in Go.
	Impl BaselineService
}

func (p *BaselineGRPCPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	proto2.RegisterBaselineServer(s, &GRPCServer{Impl: p.Impl})
	return nil
}

func (p *BaselineGRPCPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCClient{client: proto2.NewBaselineClient(c)}, nil
}

// GRPCClient is an implementation of Baseline that talks over RPC.
type GRPCClient struct {
	client proto2.BaselineClient
}

func (m *GRPCClient) GetMatrixProfileResults(situationID int64, situationInstanceID int64, ti time.Time, expressionFacts map[string]float64) (map[string]MatrixProfileResult, error) {

	results := make(map[string]MatrixProfileResult, 0)

	resp, err := m.client.GetMatrixProfileResults(context.Background(), &proto2.MatrixProfileResultRequest{
		SituationId:         situationID,
		SituationInstanceId: situationInstanceID,
		Time:                ti.Format(timeLayout),
		ExpressionFacts:     expressionFacts,
	})
	if err != nil {
		return results, err
	}

	for k, v := range resp.Values {
		// An unparseable timestamp must not drop the whole entry: business rules address these
		// results by path, and a missing key makes the condition false with no trace. Keep the
		// status and the scores, which are what rules actually test, and leave the time zeroed.
		bestMatchTime, err := time.Parse(timeLayout, v.GetBestMatchTime())
		if err != nil {
			zap.L().Warn("parse matrix profile best match time", zap.String("definition", k), zap.Error(err))
			bestMatchTime = time.Time{}
		}
		results[k] = MatrixProfileResult{
			Status:                   v.Status,
			MatchingScore:            v.MatchingScore,
			MatchingScoreTheoretical: v.MatchingScoreTheoretical,
			BestMatchTime:            bestMatchTime,
		}
	}

	return results, nil
}

type GRPCServer struct {
	// This is the real implementation
	Impl BaselineService
	proto2.UnimplementedBaselineServer
}

func (m *GRPCServer) GetMatrixProfileResults(ctx context.Context, req *proto2.MatrixProfileResultRequest) (*proto2.MatrixProfileResults, error) {
	ti, err := time.Parse(timeLayout, req.Time)
	if err != nil {
		return nil, err
	}

	results, err := m.Impl.GetMatrixProfileResults(req.SituationId, req.SituationInstanceId, ti, req.ExpressionFacts)

	values := make(map[string]*proto2.MatrixProfileResult, 0)
	for k, v := range results {
		values[k] = &proto2.MatrixProfileResult{
			Status:                   v.Status,
			MatchingScore:            v.MatchingScore,
			MatchingScoreTheoretical: v.MatchingScoreTheoretical,
			BestMatchTime:            v.BestMatchTime.Format(timeLayout),
		}
	}

	return &proto2.MatrixProfileResults{Values: values}, err
}
