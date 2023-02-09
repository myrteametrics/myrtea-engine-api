package coordinator

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/myrteametrics/myrtea-sdk/v4/modeler"
	"github.com/myrteametrics/myrtea-sdk/v4/models"

	"github.com/myrteametrics/myrtea-sdk/v4/elasticsearch"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// LogicalIndex abstracts a group a technical elasticsearch indices, which are accessibles with specific aliases
type LogicalIndexTimeBased struct {
	Initialized bool
	Name        string
	Cron        *cron.Cron
	Executor    *elasticsearch.EsExecutor
	Model       modeler.Model
	LiveIndices []string
	mu          sync.RWMutex
}

func NewLogicalIndexTimeBased(instanceName string, model modeler.Model, executor *elasticsearch.EsExecutor) (*LogicalIndexTimeBased, error) {

	logicalIndexName := fmt.Sprintf("%s-%s", instanceName, model.Name)

	zap.L().Info("Initialize logicalIndex (LogicalIndexTimeBased)", zap.String("name", logicalIndexName), zap.String("model", model.Name), zap.Any("options", model.ElasticsearchOptions))

	if model.ElasticsearchOptions.Rollmode != "timebased" {
		return nil, errors.New("invalid rollmode for this logicalIndex type")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	templateName := fmt.Sprintf("template-%s", logicalIndexName)
	templateExists, err := executor.Client.IndexTemplateExists(templateName).Do(ctx)
	if err != nil {
		zap.L().Error("IndexTemplateExists()", zap.Error(err))
		return nil, err
	}
	if !templateExists {
		indexPatern := fmt.Sprintf("%s-*", logicalIndexName)
		templateBody := models.NewTemplate([]string{indexPatern}, model.ToElasticsearchMappingProperties(), model.ElasticsearchOptions.AdvancedSettings)
		err := executor.PutTemplate(ctx, templateName, templateBody)
		if err != nil {
			zap.L().Error("PutTemplate()", zap.Error(err))
			return nil, err
		}
	}

	logicalIndex := &LogicalIndexTimeBased{
		Initialized: false,
		Name:        logicalIndexName,
		Cron:        nil,
		Executor:    executor,
		Model:       model,
		LiveIndices: make([]string, 0),
	}

	logicalIndex.FetchIndices()

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

	tsStart := time.Now().Add(time.Duration(logicalIndex.Model.ElasticsearchOptions.PurgeMaxConcurrentIndices) * -1 * 24 * time.Hour)
	indexStart := fmt.Sprintf("%s-%s", logicalIndex.Name, tsStart.Format("2006-01-02"))

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	catIndicesResponse, err := logicalIndex.Executor.Client.CatIndices().Index(fmt.Sprintf("%s-*", logicalIndex.Name)).Columns("index").Do(ctx)
	if err != nil {
		zap.L().Error("elasticsearch CatIndices", zap.Error(err))
		return
	}

	indices := make([]string, 0)
	for _, index := range catIndicesResponse {
		if index.Index < indexStart { // purge selection condition
			indices = append(indices, index.Index)
		}
	}
	sort.Strings(indices)

	zap.L().Info("Purging indices older than", zap.String("indexStart", indexStart), zap.Strings("indices", indices))

	if len(indices) > 0 {
		ctxDelete, cancelDelete := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancelDelete()
		_, err = logicalIndex.Executor.Client.DeleteIndex(indices...).Do(ctxDelete)
		if err != nil {
			zap.L().Warn("Delete index", zap.Error(err))
		}
	}

	logicalIndex.FetchIndices()

}

func (logicalIndex *LogicalIndexTimeBased) FetchIndices() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	catIndicesResponse, err := logicalIndex.Executor.Client.CatIndices().Index(fmt.Sprintf("%s-*", logicalIndex.Name)).Columns("index").Do(ctx)
	if err != nil {
		zap.L().Error("elasticsearch CatIndices", zap.Error(err))
		return
	}

	indices := make([]string, 0)
	for _, index := range catIndicesResponse {
		indices = append(indices, index.Index)
	}
	sort.Strings(indices)

	logicalIndex.mu.Lock()
	logicalIndex.LiveIndices = indices
	logicalIndex.mu.Unlock()
}

func (logicalIndex *LogicalIndexTimeBased) FindIndices(t time.Time, depthDays int64) ([]string, error) {
	tsStart := t.Add(time.Duration(depthDays) * -1 * 24 * time.Hour)
	indexEnd := fmt.Sprintf("%s-%s", logicalIndex.Name, t.Format("2006-01-02"))
	indexStart := fmt.Sprintf("%s-%s", logicalIndex.Name, tsStart.Format("2006-01-02"))

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
