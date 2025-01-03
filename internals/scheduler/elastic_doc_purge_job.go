package scheduler

import (
	"errors"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/fact"
	"github.com/myrteametrics/myrtea-sdk/v5/engine"
	"go.uber.org/zap"
	"time"
)

type ElasticDocPurgeJob struct {
	FactIds    []int64 `json:"factIds"`
	ScheduleID int64   `json:"-"`
}

func (job ElasticDocPurgeJob) IsValid() (bool, error) {
	if job.FactIds == nil || len(job.FactIds) <= 0 {
		return false, errors.New("missing FactIds")
	}
	return true, nil
}

func (job ElasticDocPurgeJob) Run() {

	if S().ExistingRunningJob(job.ScheduleID) {
		zap.L().Info("Skipping Elastic document purge job because last execution is still running", zap.Int64s("ids", job.FactIds))
		return
	}
	S().AddRunningJob(job.ScheduleID)

	zap.L().Info("Delete Elastic document job started", zap.Int64s("ids", job.FactIds))

	t := time.Now().Truncate(1 * time.Second).UTC()

	PurgeElasticDocs(t, job.FactIds)

	zap.L().Info("Elastic document purge job ended", zap.Int64("id Schedule", job.ScheduleID))
	S().RemoveRunningJob(job.ScheduleID)

}

func PurgeElasticDocs(t time.Time, factIds []int64) {
	zap.L().Info("Starting Elastic document purge", zap.Time("timestamp", t), zap.Int("number_of_facts", len(factIds)))

	for _, factId := range factIds {
		zap.L().Debug("Processing fact", zap.Int64("factId", factId))

		// Retrieve the Fact
		f, found, err := fact.R().Get(factId)
		if err != nil {
			zap.L().Error("Error retrieving the fact; skipping deletion",
				zap.Int64("factId", factId),
				zap.Error(err))
			continue
		}

		if !found {
			zap.L().Warn("Fact does not exist; skipping deletion", zap.Int64("factId", factId))
			continue
		}

		if f.Intent.Operator != engine.Delete {
			zap.L().Warn("Fact is not a delete fact; skipping deletion", zap.Int64("factId", factId))
		}

		// Execute deletion
		_, err = fact.ExecuteFactDeleteQuery(t, f)
		if err != nil {
			zap.L().Error("Error during fact deletion",
				zap.Int64("factId", f.ID),
				zap.Any("fact", f),
				zap.Error(err))
			continue
		}

		zap.L().Info("Fact successfully deleted from Elastic", zap.Int64("factId", f.ID))
	}

	zap.L().Info("Elastic document purge completed", zap.Time("timestamp", t), zap.Int("processed_facts", len(factIds)))
}
