package coordinator

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/elastic/go-elasticsearch/v8/typedapi/indices/rollover"
	"github.com/elastic/go-elasticsearch/v8/typedapi/indices/updatealiases"
	"github.com/elastic/go-elasticsearch/v8/typedapi/some"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/myrteametrics/myrtea-sdk/v4/elasticsearchv8"
	"github.com/myrteametrics/myrtea-sdk/v4/index"
	"github.com/myrteametrics/myrtea-sdk/v4/modeler"

	"github.com/myrteametrics/myrtea-sdk/v4/postgres"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// LogicalIndexCronV8 abstracts a group a technical elasticsearchv6 indices, which are accessibles with specific aliases
type LogicalIndexCronV8 struct {
	Initialized bool
	Name        string
	Cron        *cron.Cron
	Model       modeler.Model
	mu          sync.RWMutex
}

func NewLogicalIndexCronV8(instanceName string, model modeler.Model) (*LogicalIndexCronV8, error) {

	logicalIndexName := fmt.Sprintf("%s-%s", instanceName, model.Name)

	zap.L().Info("Initialize logicalIndex (LogicalIndexCronV8)", zap.String("name", logicalIndexName), zap.String("model", model.Name), zap.Any("options", model.ElasticsearchOptions))

	if model.ElasticsearchOptions.Rollmode != "cron" {
		return nil, errors.New("invalid rollmode for this logicalIndex type")
	}

	logicalIndex := &LogicalIndexCronV8{
		Initialized: false,
		Name:        logicalIndexName,
		Cron:        nil,
		Model:       model,
	}

	ctx := context.Background()
	indexPatern := fmt.Sprintf("%s-active-*", logicalIndexName)
	//exists, err := elasticsearchv8.C().Indices.Exists(indexPatern).Do(ctx)
	//exists, err := elasticsearchv8.C().Indices.ExistsTemplate()
	templateName := fmt.Sprintf("template-%s", logicalIndexName)
	templateExists, err := elasticsearchv8.C().Indices.ExistsTemplate(templateName).IsSuccess(ctx)
	if err != nil {
		zap.L().Error("IndexTemplateExists()", zap.Error(err))
		return nil, err
	}

	if !templateExists {

		//templateName := fmt.Sprintf("template-%s", logicalIndexName)
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
		defer cancel()

		logicalIndex.putTemplate(templateName, indexPatern, model)

		_, err := elasticsearchv8.C().Indices.Create(logicalIndexName + "-active-000001").Do(ctx)
		if err != nil {
			zap.L().Error("elasticsearchv8.C().PutIndex()", zap.Error(err))
			return nil, err
		}

		_, err = elasticsearchv8.C().Indices.PutAlias(logicalIndexName+"-active-*", logicalIndexName+"-current").Do(ctx)
		if err != nil {
			zap.L().Error("elasticsearchv8.C().PutAlias()", zap.Error(err))
			return nil, err
		}

		if model.ElasticsearchOptions.PatchAliasMaxIndices > 0 {
			_, err = elasticsearchv8.C().Indices.PutAlias(logicalIndexName+"-*", logicalIndexName+"-patch").Do(ctx)
			if err != nil {
				zap.L().Error("elasticsearchv8.C().PutAlias()", zap.Error(err))
				return nil, err
			}
		}

		// Adding search alias on all active index
		_, err = elasticsearchv8.C().Indices.PutAlias(logicalIndexName+"-*", logicalIndexName+"-search").Do(ctx)
		if err != nil {
			zap.L().Error("elasticsearchv8.C().PutAlias()", zap.Error(err))
			return nil, err
		}

		err = logicalIndex.persistTechnicalIndex(logicalIndexName+"-active-000001", time.Now().UTC())
		if err != nil {
			zap.L().Error("Could not persist technical index data", zap.Error(err))
		}
	}

	c := cron.New()
	_, err = c.AddFunc(model.ElasticsearchOptions.Rollcron, logicalIndex.rollover)
	if err != nil {
		zap.L().Error("Cron add function logicalIndex.updateAliases", zap.Error(err))
		return nil, err
	}
	logicalIndex.Cron = c
	zap.L().Info("Cron started", zap.String("logicalIndex", logicalIndexName), zap.String("cron", logicalIndex.Model.ElasticsearchOptions.Rollcron))

	logicalIndex.Initialized = true

	return logicalIndex, nil
}

func (logicalIndex *LogicalIndexCronV8) putTemplate(name string, indexPatern string, model modeler.Model) {
	req := elasticsearchv8.NewPutTemplateRequestV8([]string{indexPatern}, model)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	response, err := elasticsearchv8.C().Indices.PutTemplate(name).Request(req).Do(ctx)
	if err != nil {
		zap.L().Error("PutTemplate", zap.Error(err))
	}
	if !response.Acknowledged {
		zap.L().Error("PutTemplate failed")
	}
}

func (logicalIndex *LogicalIndexCronV8) GetCron() *cron.Cron {
	return logicalIndex.Cron
}

// FindIndicesWithDynamicDepth search in indices referential every indices between two dates (calculated using current time and depth)
func (logicalIndex *LogicalIndexCronV8) FindIndices(t time.Time, depthDays int64) ([]string, error) {
	if depthDays == 0 {
		return []string{fmt.Sprintf("%s-%s", logicalIndex.Name, index.All)}, nil
	}

	query := "select technical from elasticsearch_indices_v1 where logical = :logical AND creation_date BETWEEN :mindate AND :maxdate"
	params := map[string]interface{}{
		"logical": logicalIndex.Name,
		"mindate": t.Add(-1 * time.Duration(depthDays+1) * 24 * time.Hour),
		"maxdate": t,
	}
	rows, err := postgres.DB().NamedQuery(query, params)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	indices := make([]string, 0)
	for rows.Next() {
		var name string
		err := rows.Scan(&name)
		if err != nil {
			return nil, err
		}
		indices = append(indices, name)
	}

	return indices, nil
}

func (logicalIndex *LogicalIndexCronV8) rollover() {
	ctx := context.Background()

	// Using rollover API to manage alias swap and indices names
	// TODO: Change this dirty abuse of rollover API (triggered every day "max_docx = 0")
	req := rollover.NewRequest()
	// req.Conditions = types.NewRolloverConditions()
	// req.Conditions.MaxAge = "24 h"
	// req.Conditions.MaxDocs = some.Int64(0)

	resRollover, err := elasticsearchv8.C().Indices.Rollover(logicalIndex.Name + "-current").Request(req).Do(ctx)
	if err != nil {
		zap.L().Error("RollOverV2", zap.Error(err))
		return
	}

	// Roll patch alias
	alias := logicalIndex.Name + "-patch"
	resAlias, err := elasticsearchv8.C().Indices.GetAlias().Name(alias).Do(ctx)
	if err != nil {
		zap.L().Error("GetIndicesByAlias", zap.Error(err))
		return
	}
	patchIndices := make([]string, 0)
	for indexName := range resAlias {
		patchIndices = append(patchIndices, indexName)
	}
	sort.Strings(patchIndices)

	if len(patchIndices) <= logicalIndex.Model.ElasticsearchOptions.PatchAliasMaxIndices {
		updateAliasesRequest := updatealiases.NewRequest()

		if len(patchIndices) >= logicalIndex.Model.ElasticsearchOptions.PatchAliasMaxIndices {
			updateAliasesRequest.Actions = append(updateAliasesRequest.Actions,
				types.IndicesAction{Remove: &types.RemoveAction{Index: some.String(patchIndices[0]), Alias: some.String(alias)}},
			)
		}
		updateAliasesRequest.Actions = append(updateAliasesRequest.Actions,
			types.IndicesAction{Add: &types.AddAction{Index: some.String(resRollover.NewIndex), Alias: some.String(alias)}},
		)

		updateAliasesResponse, err := elasticsearchv8.C().Indices.UpdateAliases().Request(updateAliasesRequest).Do(ctx)
		if err != nil {
			zap.L().Error("Putting alias", zap.Error(err), zap.String("index", resRollover.NewIndex),
				zap.String("alias", alias))
			return
		}
		if !updateAliasesResponse.Acknowledged {
			err := errors.New("es API return false acknowledged")
			zap.L().Error("Putting alias", zap.Error(err), zap.String("index", resRollover.NewIndex),
				zap.String("alias", alias))
			return
		}

		zap.L().Info("Patch aliases rollover done", zap.String("alias", alias))
	}

	// Adding search alias on the newly created active index
	// Includes every active + inactive indices
	// TODO: Remove this step when template are reworked (and include this search alias)
	_, err = elasticsearchv8.C().Indices.PutAlias(logicalIndex.Name+"-*", logicalIndex.Name+"-search").Do(ctx)
	if err != nil {
		zap.L().Error("Putting alias", zap.Error(err), zap.String("index", logicalIndex.Name+"-*"),
			zap.String("alias", logicalIndex.Name+"-search"))
		return
	}

	err = logicalIndex.persistTechnicalIndex(resRollover.NewIndex, time.Now().UTC())
	if err != nil {
		zap.L().Error("Could not persist technical index data", zap.Error(err))
	}

	// Purge outdated indices
	if logicalIndex.Model.ElasticsearchOptions.EnablePurge {
		resAlias, err := elasticsearchv8.C().Indices.GetAlias().Name(logicalIndex.Name + "-search").Do(ctx)
		if err != nil {
			zap.L().Error("GetIndicesByAlias", zap.Error(err))
			return
		}
		searchIndices := make([]string, 0)
		for indexName := range resAlias {
			searchIndices = append(searchIndices, indexName)
		}
		sort.Strings(searchIndices)

		if len(searchIndices) > logicalIndex.Model.ElasticsearchOptions.PurgeMaxConcurrentIndices {
			toDelete := searchIndices[0] // oldest indices

			_, err := elasticsearchv8.C().Indices.Delete(toDelete).Do(ctx)
			if err != nil {
				zap.L().Error("DeleteIndex", zap.Error(err), zap.String("index", searchIndices[0]))
				return
			}
			err = logicalIndex.purgeTechnicalIndex(toDelete)
			if err != nil {
				zap.L().Error("Could not persist technical index data", zap.Error(err))
			}
			zap.L().Info("Index purged", zap.String("name", toDelete))
		}
	}
}

func (logicalIndex *LogicalIndexCronV8) persistTechnicalIndex(newIndex string, t time.Time) error {
	if postgres.DB() == nil {
		return errors.New("postgresql Client not initialized")
	}

	query := `INSERT INTO elasticsearch_indices_v1 (id, logical, technical, creation_date)
		VALUES (DEFAULT, :logical, :technical, :creation_date);`
	params := map[string]interface{}{
		"logical":       logicalIndex.Name,
		"technical":     newIndex,
		"creation_date": t,
	}

	_, err := postgres.DB().NamedExec(query, params)
	if err != nil {
		return err
	}

	return nil
}

func (logicalIndex *LogicalIndexCronV8) purgeTechnicalIndex(technicalIndex string) error {
	if postgres.DB() == nil {
		return errors.New("postgresql Client not initialized")
	}

	query := `DELETE FROM elasticsearch_indices_v1 where logical = :logical AND technical = :technical`
	params := map[string]interface{}{
		"logical":   logicalIndex.Name,
		"technical": technicalIndex,
	}

	_, err := postgres.DB().NamedExec(query, params)
	if err != nil {
		return err
	}

	return nil
}
