package ingester

import (
	"errors"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/evaluator"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/metrics"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/scheduler"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/tasker"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// AggregateIngester is a component which process scheduler.ExternalAggregate
type AggregateIngester struct {
	Data                              chan []scheduler.ExternalAggregate
	metricAggregateIngesterQueueGauge stdprometheus.Gauge
}

// NewAggregateIngester returns a pointer to a new AggregateIngester instance
func NewAggregateIngester() *AggregateIngester {
	var aggregateIngesterGauge = stdprometheus.NewGauge(stdprometheus.GaugeOpts{
		Namespace:   metrics.MetricNamespace,
		ConstLabels: metrics.MetricPrometheusLabels,
		Name:        "aggregateingester_queue",
		Help:        "this is the help string for aggregateingester_queue",
	})

	// Register metrics
	stdprometheus.MustRegister(aggregateIngesterGauge)
	aggregateIngesterGauge.Set(0)

	return &AggregateIngester{
		Data:                              make(chan []scheduler.ExternalAggregate, viper.GetInt("AGGREGATEINGESTER_QUEUE_BUFFER_SIZE")),
		metricAggregateIngesterQueueGauge: aggregateIngesterGauge,
	}
}

// Run is the main routine of a TypeIngester instance
func (ingester *AggregateIngester) Run() {
	zap.L().Info("Starting AggregateIngester")

	for ir := range ingester.Data {
		zap.L().Debug("Received ExternalAggregate", zap.Int("ExternalAggregate items count", len(ir)))

		err := HandleAggregates(ir)
		if err != nil {
			zap.L().Error("AggregateIngester: could not handle aggregates", zap.Error(err))
		}

		// Update queue gauge
		ingester.metricAggregateIngesterQueueGauge.Set(float64(len(ingester.Data)))
	}

}

// Ingest process an array of scheduler.ExternalAggregate
func (ingester *AggregateIngester) Ingest(aggregates []scheduler.ExternalAggregate) error {
	dataLen := len(ingester.Data)

	// Check for channel overloading
	if dataLen+1 >= cap(ingester.Data) {
		zap.L().Debug("Buffered channel would be overloaded with incoming bulkIngestRequest")
		ingester.metricAggregateIngesterQueueGauge.Set(float64(dataLen))
		return errors.New("channel overload")
	}

	ingester.Data <- aggregates

	return nil
}

// HandleAggregates process a slice of ExternalAggregates and trigger all standard fact-situation-rule process
func HandleAggregates(aggregates []scheduler.ExternalAggregate) error {
	localRuleEngine, err := evaluator.BuildLocalRuleEngine("external-aggs")
	if err != nil {
		zap.L().Error("BuildLocalRuleEngine", zap.Error(err))
		return err
	}

	situationsToUpdate, err := scheduler.ReceiveAndPersistFacts(aggregates)
	if err != nil {
		zap.L().Error("ReceiveAndPersistFacts", zap.Error(err))
		return err
	}

	taskBatchs, err := scheduler.CalculateAndPersistSituations(localRuleEngine, situationsToUpdate)
	if err != nil {
		zap.L().Error("CalculateAndPersistSituations", zap.Error(err))
		return err
	}

	tasker.T().BatchReceiver <- taskBatchs

	return nil
}
