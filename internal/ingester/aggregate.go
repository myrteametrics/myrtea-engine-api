package ingester

import (
	"errors"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/evaluator"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/metrics"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/scheduler"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/tasker"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// AggregateIngester is a component which process scheduler.ExternalAggregate
type AggregateIngester struct {
	data             chan []scheduler.ExternalAggregate
	metricQueueGauge *stdprometheus.Gauge
	running          bool
}

var (
	_aggregateIngesterGauge = _newRegisteredGauge()
)

func _newRegisteredGauge() *stdprometheus.Gauge {
	var gauge = stdprometheus.NewGauge(stdprometheus.GaugeOpts{
		Namespace:   metrics.MetricNamespace,
		ConstLabels: metrics.MetricPrometheusLabels,
		Name:        "aggregateingester_queue",
		Help:        "this is the help string for aggregateingester_queue",
	})

	// Register metrics
	stdprometheus.MustRegister(gauge)
	gauge.Set(0)

	return &gauge
}

// NewAggregateIngester returns a pointer to a new AggregateIngester instance
func NewAggregateIngester() *AggregateIngester {
	return &AggregateIngester{
		data:             make(chan []scheduler.ExternalAggregate, viper.GetInt("AGGREGATEINGESTER_QUEUE_BUFFER_SIZE")),
		metricQueueGauge: _aggregateIngesterGauge,
		running:          false,
	}
}

// Run is the main routine of a TypeIngester instance
func (ai *AggregateIngester) Run() {
	zap.L().Info("Starting AggregateIngester")

	for ir := range ai.data {
		zap.L().Debug("Received ExternalAggregate", zap.Int("ExternalAggregate items count", len(ir)))

		err := HandleAggregates(ir)
		if err != nil {
			zap.L().Error("AggregateIngester: could not handle aggregates", zap.Error(err))
		}

		// Update queue gauge
		(*ai.metricQueueGauge).Set(float64(len(ai.data)))
	}

}

// Ingest process an array of scheduler.ExternalAggregate
func (ai *AggregateIngester) Ingest(aggregates []scheduler.ExternalAggregate) error {
	dataLen := len(ai.data)

	// Start ingester if not running
	if !ai.running {
		go ai.Run()
		ai.running = true
	}

	zap.L().Debug("Ingesting data", zap.Any("aggregates", aggregates))

	// Check for channel overloading
	if dataLen+1 >= cap(ai.data) {
		zap.L().Debug("Buffered channel would be overloaded with incoming bulkIngestRequest")
		(*ai.metricQueueGauge).Set(float64(dataLen))
		return errors.New("channel overload")
	}

	ai.data <- aggregates

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
