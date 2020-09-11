package baseline

import (
	"time"

	"github.com/hashicorp/go-plugin"
	"github.com/myrteametrics/myrtea-engine-api/v4/plugins/baseline/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const timeLayout = "2006-01-02T15:04:05.000Z07:00"

type BaselineGRPCPlugin struct {
	// GRPCPlugin must still implement the Plugin interface
	plugin.Plugin
	// Concrete implementation, written in Go. This is only used for plugins that are written in Go.
	Impl Baseline
}

func (p *BaselineGRPCPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	proto.RegisterBaselineServer(s, &GRPCServer{Impl: p.Impl})
	return nil
}

func (p *BaselineGRPCPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCClient{client: proto.NewBaselineClient(c)}, nil
}

// GRPCClient is an implementation of Baseline that talks over RPC.
type GRPCClient struct {
	client proto.BaselineClient
}

func (m *GRPCClient) GetBaselineValues(id int64, factID int64, situationID int64, situationInstanceID int64, ti time.Time) (map[string]BaselineValue, error) {

	baselineValues := make(map[string]BaselineValue, 0)

	resp, err := m.client.GetBaselineValues(context.Background(), &proto.BaselineValueRequest{
		Id:                  id,
		FactId:              factID,
		SituationId:         situationID,
		SituationInstanceId: situationInstanceID,
		Time:                ti.Format(timeLayout),
	})
	if err != nil {
		return baselineValues, err
	}

	for k, v := range resp.Values {
		vTime, err := time.Parse(v.GetTime(), timeLayout)
		if err != nil {
			continue
		}
		baselineValues[k] = BaselineValue{
			Time:       vTime,
			Value:      v.Value,
			ValueLower: v.ValueLower,
			ValueUpper: v.ValueUpper,
			Avg:        v.Avg,
			Std:        v.Std,
			Median:     v.Median,
		}

	}

	return baselineValues, nil
}

type GRPCServer struct {
	// This is the real implementation
	Impl Baseline
}

func (m *GRPCServer) GetBaselineValues(ctx context.Context, req *proto.BaselineValueRequest) (*proto.BaselineValues, error) {
	ti, err := time.Parse(timeLayout, req.Time)
	if err != nil {
		return nil, err
	}

	values, err := m.Impl.GetBaselineValues(req.Id, req.FactId, req.SituationId, req.SituationInstanceId, ti)

	baselineValues := make(map[string]*proto.BaselineValue, 0)
	for k, v := range values {
		baselineValues[k] = &proto.BaselineValue{
			Time:       v.Time.Format(timeLayout),
			Value:      v.Value,
			ValueLower: v.ValueLower,
			ValueUpper: v.ValueUpper,
			Avg:        v.Avg,
			Std:        v.Std,
			Median:     v.Median,
		}
	}

	return &proto.BaselineValues{Values: baselineValues}, err
}
