package baseline

import (
	"time"

	proto2 "github.com/myrteametrics/myrtea-engine-api/v5/pkg/plugins/baseline/proto"
	"google.golang.org/protobuf/types/known/emptypb"

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

func (m *GRPCClient) GetBaselineValues(id int64, factID int64, situationID int64, situationInstanceID int64, ti time.Time) (map[string]BaselineValue, error) {

	baselineValues := make(map[string]BaselineValue, 0)

	resp, err := m.client.GetBaselineValues(context.Background(), &proto2.BaselineValueRequest{
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
		vTime, err := time.Parse(timeLayout, v.GetTime())
		if err != nil {
			zap.L().Warn("parse baseline ti", zap.Error(err))
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

func (m *GRPCClient) BuildBaselineValues(baselineID int64) error {

	_, err := m.client.BuildBaselineValues(context.Background(), &proto2.BuildBaselineRequest{
		Id: baselineID,
	})
	if err != nil {
		return err
	}
	return nil
}

func (m *GRPCClient) GetMatrixProfileResults(situationID int64, situationInstanceID int64, ti time.Time) (map[string]MatrixProfileResult, error) {

	results := make(map[string]MatrixProfileResult, 0)

	resp, err := m.client.GetMatrixProfileResults(context.Background(), &proto2.MatrixProfileResultRequest{
		SituationId:         situationID,
		SituationInstanceId: situationInstanceID,
		Time:                ti.Format(timeLayout),
	})
	if err != nil {
		return results, err
	}

	for k, v := range resp.Values {
		bestMatchTime, err := time.Parse(timeLayout, v.GetBestMatchTime())
		if err != nil {
			zap.L().Warn("parse matrix profile best match time", zap.Error(err))
			continue
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

func (m *GRPCServer) GetBaselineValues(ctx context.Context, req *proto2.BaselineValueRequest) (*proto2.BaselineValues, error) {
	ti, err := time.Parse(timeLayout, req.Time)
	if err != nil {
		return nil, err
	}

	values, err := m.Impl.GetBaselineValues(req.Id, req.FactId, req.SituationId, req.SituationInstanceId, ti)

	baselineValues := make(map[string]*proto2.BaselineValue, 0)
	for k, v := range values {
		baselineValues[k] = &proto2.BaselineValue{
			Time:       v.Time.Format(timeLayout),
			Value:      v.Value,
			ValueLower: v.ValueLower,
			ValueUpper: v.ValueUpper,
			Avg:        v.Avg,
			Std:        v.Std,
			Median:     v.Median,
		}
	}

	return &proto2.BaselineValues{Values: baselineValues}, err
}

func (m *GRPCServer) BuildBaselineValues(ctx context.Context, req *proto2.BuildBaselineRequest) (*emptypb.Empty, error) {
	err := m.Impl.BuildBaselineValues(req.Id)
	return &emptypb.Empty{}, err
}

func (m *GRPCServer) GetMatrixProfileResults(ctx context.Context, req *proto2.MatrixProfileResultRequest) (*proto2.MatrixProfileResults, error) {
	ti, err := time.Parse(timeLayout, req.Time)
	if err != nil {
		return nil, err
	}

	results, err := m.Impl.GetMatrixProfileResults(req.SituationId, req.SituationInstanceId, ti)

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
