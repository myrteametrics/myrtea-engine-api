package coordinator

import (
	"context"
	"errors"
	"fmt"
	"github.com/myrteametrics/myrtea-sdk/v5/elasticsearch"
	"sort"
	"sync"
	"time"

	"github.com/myrteametrics/myrtea-sdk/v5/modeler"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// LogicalIndexTimeBased abstracts a group a technical elasticsearchv8 indices, which are accessibles with specific aliases
type LogicalIndexTimeBased struct {
	Initialized bool
	Name        string
	Cron        *cron.Cron
	Model       modeler.Model
	LiveIndices []string
	mu          sync.RWMutex
}

func NewLogicalIndexTimeBased(instanceName string, model modeler.Model) (*LogicalIndexTimeBased, error) {

	logicalIndexName := fmt.Sprintf("%s-%s", instanceName, model.Name)

	zap.L().Info("Initialize logicalIndex (LogicalIndexTimeBased)", zap.String("name", logicalIndexName), zap.String("model", model.Name), zap.Any("options", model.ElasticsearchOptions))

	if model.ElasticsearchOptions.Rollmode.Type != modeler.RollmodeTimeBased {
		return nil, errors.New("invalid rollmode for this logicalIndex type")
	}

	templateName := fmt.Sprintf("template-%s", logicalIndexName)

	logicalIndex := &LogicalIndexTimeBased{
		Initialized: false,
		Name:        logicalIndexName,
		Cron:        nil,
		Model:       model,
		LiveIndices: make([]string, 0),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	templateExists, err := elasticsearch.C().Indices.ExistsTemplate(templateName).IsSuccess(ctx)
	if err != nil {
		zap.L().Error("IndexTemplateExists()", zap.Error(err))
		return nil, err
	}
	if !templateExists {
		zap.L().Info("template doesn't exists, creating it", zap.String("logicalIndexName", logicalIndexName))
		indexPatern := fmt.Sprintf("%s-*", logicalIndexName)
		logicalIndex.putTemplate(templateName, indexPatern, model)
	}

	logicalIndex.FetchIndices()
	zap.L().Info("LogicalIndex initialized", zap.String("logicalIndex", logicalIndex.Name))

	c := cron.New()
	_, err = c.AddFunc("*/30 * * * *", logicalIndex.FetchIndices)
	if err != nil {
		zap.L().Error("Cron add function logicalIndex.updateAliases", zap.Error(err))
		return nil, err
	}

	if logicalIndex.Model.ElasticsearchOptions.EnablePurge {
		_, err = c.AddFunc(logicalIndex.Model.ElasticsearchOptions.Rollcron, logicalIndex.purge)
		if err != nil {
			zap.L().Error("Cron add function logicalIndex.updateAliases", zap.Error(err))
			return nil, err
		}
	}

	logicalIndex.Cron = c
	zap.L().Info("Cron started", zap.String("logicalIndex", logicalIndex.Name), zap.String("cron", logicalIndex.Model.ElasticsearchOptions.Rollcron))

	return logicalIndex, nil
}

func (logicalIndex *LogicalIndexTimeBased) GetCron() *cron.Cron {
	return logicalIndex.Cron
}

func (logicalIndex *LogicalIndexTimeBased) purge() {
	if !logicalIndex.Model.ElasticsearchOptions.EnablePurge {
		return
	}

	var indexStart string
	var tsStart time.Time

	if logicalIndex.Model.ElasticsearchOptions.Rollmode.Timebased.Interval == modeler.Daily {
		// Daily mode: calculation based on a number of days
		tsStart = time.Now().Add(time.Duration(logicalIndex.Model.ElasticsearchOptions.PurgeMaxConcurrentIndices) * -1 * 24 * time.Hour)
		indexStart = fmt.Sprintf("%s-%s", logicalIndex.Name, tsStart.Format("2006-01-02"))
	} else if logicalIndex.Model.ElasticsearchOptions.Rollmode.Timebased.Interval == modeler.Monthly {
		// Monthly mode: calculation based on a number of months
		tsStart = time.Now().AddDate(0, -logicalIndex.Model.ElasticsearchOptions.PurgeMaxConcurrentIndices, 0)
		indexStart = fmt.Sprintf("%s-%s", logicalIndex.Name, tsStart.Format("2006-01"))
	}

	allIndices := logicalIndex.GetAllIndices()
	indices := make([]string, 0)
	for _, index := range allIndices {
		if index < indexStart { // purge selection condition
			indices = append(indices, index)
		}
	}
	sort.Strings(indices)

	zap.L().Info("Purging indices older than", zap.String("indexStart", indexStart), zap.Strings("indices", indices))

	if len(indices) > 0 {
		for _, index := range indices {
			logicalIndex.deleteIndex(index)
		}
	}

	logicalIndex.FetchIndices()
}

func (logicalIndex *LogicalIndexTimeBased) putTemplate(name string, indexPatern string, model modeler.Model) {
	req := elasticsearch.NewPutTemplateRequestV8([]string{indexPatern}, model)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	response, err := elasticsearch.C().Indices.PutTemplate(name).Request(req).Do(ctx)
	if err != nil {
		zap.L().Error("PutTemplate", zap.Error(err))
	}
	if !response.Acknowledged {
		zap.L().Error("PutTemplate failed")
	}
}

func (logicalIndex *LogicalIndexTimeBased) deleteIndex(index string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	success, err := elasticsearch.C().Indices.Delete(index).IsSuccess(ctx)
	if err != nil {
		zap.L().Warn("Delete index failed", zap.Error(err))
	}
	if !success {
		zap.L().Warn("Delete index failed")
	}
}

func (logicalIndex *LogicalIndexTimeBased) FetchIndices() {
	indices := logicalIndex.GetAllIndices()

	logicalIndex.mu.Lock()
	logicalIndex.LiveIndices = indices
	logicalIndex.mu.Unlock()
}

func (logicalIndex *LogicalIndexTimeBased) GetAllIndices() []string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	response, err := elasticsearch.C().Cat.Indices().Index(fmt.Sprintf("%s-*", logicalIndex.Name)).Do(ctx)
	if err != nil {
		zap.L().Error("elasticsearchv8 CatIndices", zap.Error(err))
		return make([]string, 0)
	}

	indices := make([]string, 0)
	for _, index := range response {
		if index.Index != nil {
			indices = append(indices, *index.Index)
		}
	}
	sort.Strings(indices)
	return indices
}

func (logicalIndex *LogicalIndexTimeBased) FindIndices(t time.Time, depthDays int64) ([]string, error) {
	var indexStart, indexEnd string

	if logicalIndex.Model.ElasticsearchOptions.Rollmode.Timebased.Interval == modeler.Daily {
		tsStart := t.Add(time.Duration(-depthDays) * 24 * time.Hour)
		indexEnd = fmt.Sprintf("%s-%s", logicalIndex.Name, t.Format("2006-01-02"))
		indexStart = fmt.Sprintf("%s-%s", logicalIndex.Name, tsStart.Format("2006-01-02"))
	} else if logicalIndex.Model.ElasticsearchOptions.Rollmode.Timebased.Interval == modeler.Monthly {
		// For a monthly interval, we calculate the starting point in months
		tsStart := t.AddDate(0, int(-depthDays/30), 0)
		indexEnd = fmt.Sprintf("%s-%s", logicalIndex.Name, t.Format("2006-01"))
		indexStart = fmt.Sprintf("%s-%s", logicalIndex.Name, tsStart.Format("2006-01"))
	}

	indices := make([]string, 0)
	logicalIndex.mu.RLock()
	for _, index := range logicalIndex.LiveIndices {
		if index >= indexStart && index <= indexEnd {
			indices = append(indices, index)
		}
	}
	logicalIndex.mu.RUnlock()
	return indices, nil
}
