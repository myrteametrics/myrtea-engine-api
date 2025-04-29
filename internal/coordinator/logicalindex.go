package coordinator

import (
	"time"

	"github.com/robfig/cron/v3"
)

type LogicalIndex interface {
	GetCron() *cron.Cron
	FindIndices(t time.Time, depthDays int64) ([]string, error)
}
