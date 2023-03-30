package coordinator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/myrteametrics/myrtea-sdk/v4/elasticsearchv8"
	"github.com/myrteametrics/myrtea-sdk/v4/modeler"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// LogicalIndex abstracts a group a technical elasticsearchv8 indices, which are accessibles with specific aliases
type LogicalIndexTimeBasedV8 struct {
	Initialized bool
	Name        string
	Cron        *cron.Cron
	Model       modeler.Model
	LiveIndices []string
	mu          sync.RWMutex
}

func NewLogicalIndexTimeBasedV8(instanceName string, model modeler.Model) (*LogicalIndexTimeBasedV8, error) {

	logicalIndexName := fmt.Sprintf("%s-%s", instanceName, model.Name)

	zap.L().Info("Initialize logicalIndex (LogicalIndexTimeBasedV8)", zap.String("name", logicalIndexName), zap.String("model", model.Name), zap.Any("options", model.ElasticsearchOptions))

	if model.ElasticsearchOptions.Rollmode != "timebased" {
		return nil, errors.New("invalid rollmode for this logicalIndex type")
	}

	templateName := fmt.Sprintf("template-%s", logicalIndexName)

	logicalIndex := &LogicalIndexTimeBasedV8{
		Initialized: false,
		Name:        logicalIndexName,
		Cron:        nil,
		Model:       model,
		LiveIndices: make([]string, 0),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	templateExists, err := elasticsearchv8.C().Indices.ExistsTemplate(templateName).IsSuccess(ctx)
	if err != nil {
		zap.L().Error("IndexTemplateExists()", zap.Error(err))
		return nil, err
	}
	if !templateExists {
		indexPatern := fmt.Sprintf("%s-*", logicalIndexName)
		logicalIndex.putTemplate(templateName, indexPatern, model)
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

func (logicalIndex *LogicalIndexTimeBasedV8) GetCron() *cron.Cron {
	return logicalIndex.Cron
}

func (logicalIndex *LogicalIndexTimeBasedV8) purge() {
	if !logicalIndex.Model.ElasticsearchOptions.EnablePurge {
		return
	}

	tsStart := time.Now().Add(time.Duration(logicalIndex.Model.ElasticsearchOptions.PurgeMaxConcurrentIndices) * -1 * 24 * time.Hour)
	indexStart := fmt.Sprintf("%s-%s", logicalIndex.Name, tsStart.Format("2006-01-02"))

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

func (logicalIndex *LogicalIndexTimeBasedV8) putTemplate(name string, indexPatern string, model modeler.Model) {
	req := elasticsearchv8.NewTemplateV8([]string{indexPatern}, model)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	res, err := elasticsearchv8.C().Indices.PutTemplate(name).Request(req).Do(ctx)
	if err != nil {
		zap.L().Error("PutTemplate", zap.Error(err))
	}
	defer res.Body.Close()
	if !(res.StatusCode >= 200 && res.StatusCode < 300) {
		zap.L().Error("PutTemplate", zap.Int("statuscode", res.StatusCode))
	}
}

func (logicalIndex *LogicalIndexTimeBasedV8) deleteIndex(index string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	success, err := elasticsearchv8.C().Indices.Delete(index).IsSuccess(ctx)
	if err != nil {
		zap.L().Warn("Delete index failed", zap.Error(err))
	}
	if !success {
		zap.L().Warn("Delete index failed")
	}
}

func (logicalIndex *LogicalIndexTimeBasedV8) FetchIndices() {
	indices := logicalIndex.GetAllIndices()

	logicalIndex.mu.Lock()
	logicalIndex.LiveIndices = indices
	logicalIndex.mu.Unlock()
}

func (logicalIndex *LogicalIndexTimeBasedV8) GetAllIndices() []string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	res, err := elasticsearchv8.C().Cat.Indices().Index(fmt.Sprintf("%s-*", logicalIndex.Name)).Do(ctx)
	if err != nil {
		zap.L().Error("elasticsearchv8 CatIndices", zap.Error(err))
		return make([]string, 0)
	}
	defer res.Body.Close()

	var catIndicesResponse []struct {
		Index string `json:"index"`
	}
	err = json.NewDecoder(res.Body).Decode(&catIndicesResponse)
	if err != nil {
		zap.L().Error("elasticsearchv8 CatIndices parse body", zap.Error(err))
		return make([]string, 0)
	}

	indices := make([]string, 0)
	for _, index := range catIndicesResponse {
		indices = append(indices, index.Index)
	}
	sort.Strings(indices)
	return indices
}

func (logicalIndex *LogicalIndexTimeBasedV8) FindIndices(t time.Time, depthDays int64) ([]string, error) {
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
