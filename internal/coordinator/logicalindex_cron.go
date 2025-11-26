package coordinator

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/myrteametrics/myrtea-sdk/v5/elasticsearch"

	"github.com/elastic/go-elasticsearch/v8/typedapi/indices/rollover"
	"github.com/elastic/go-elasticsearch/v8/typedapi/indices/updatealiases"
	"github.com/elastic/go-elasticsearch/v8/typedapi/some"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/myrteametrics/myrtea-sdk/v5/index"
	"github.com/myrteametrics/myrtea-sdk/v5/modeler"

	"github.com/myrteametrics/myrtea-sdk/v5/postgres"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// LogicalIndexCron abstracts a group a technical elasticsearchv8 indices, which are accessibles with specific aliases
type LogicalIndexCron struct {
	Initialized bool
	Name        string
	Cron        *cron.Cron
	Model       modeler.Model
	mu          sync.RWMutex
}

func NewLogicalIndexCronTemplate(instanceName string, model modeler.Model, updateIfExists bool) (*LogicalIndexCron, bool, error) {
	logicalIndexName := fmt.Sprintf("%s-%s", instanceName, model.Name)

	zap.L().Info("Initialize logicalIndex (LogicalIndexCron)", zap.String("name", logicalIndexName), zap.String("model", model.Name), zap.Any("options", model.ElasticsearchOptions))

	if model.ElasticsearchOptions.Rollmode.Type != modeler.RollmodeCron {
		return nil, false, errors.New("invalid rollmode for this logicalIndex type")
	}

	logicalIndex := &LogicalIndexCron{
		Initialized: false,
		Name:        logicalIndexName,
		Cron:        nil,
		Model:       model,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	indexPattern := fmt.Sprintf("%s-*", logicalIndexName)

	// First create template if not exists or update it

	templateName := fmt.Sprintf("template-%s", logicalIndexName)
	templateExists, err := elasticsearch.C().Indices.ExistsTemplate(templateName).IsSuccess(ctx)
	if err != nil {
		zap.L().Error("IndexTemplateExists()", zap.String("templateName", templateName), zap.Error(err))
		return nil, false, err
	}
	if !templateExists {
		zap.L().Info("Creating missing template", zap.String("templateName", templateName),
			zap.String("indexPattern", indexPattern), zap.String("model", model.Name))
	} else {
		zap.L().Info("Updating missing template", zap.String("templateName", templateName),
			zap.String("indexPattern", indexPattern), zap.String("model", model.Name))
	}
	if updateIfExists || !templateExists {
		logicalIndex.putTemplate(templateName, indexPattern, model)
	}

	return logicalIndex, templateExists, nil
}

func NewLogicalIndexCron(instanceName string, model modeler.Model, updateIfExists bool) (*LogicalIndexCron, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	logicalIndex, templateExists, err := NewLogicalIndexCronTemplate(instanceName, model, updateIfExists)
	if err != nil {
		return nil, err
	}

	logicalIndexName := logicalIndex.Name

	// Then create base index name if not exists
	// Use date-based naming: logicalIndexName-YYYY-MM-0001
	now := time.Now().UTC()
	baseIndexName := fmt.Sprintf("%s-%s-0001", logicalIndexName, now.Format("2006-01"))
	baseIndexExists, err := elasticsearch.C().Indices.Exists(baseIndexName).IsSuccess(ctx)
	if err != nil {
		zap.L().Error("IndexExists()", zap.String("baseIndexName", baseIndexName), zap.Error(err))
		return nil, err
	}
	if !baseIndexExists {
		zap.L().Info("Creating missing index", zap.String("baseIndexName", baseIndexName),
			zap.String("model", model.Name))
		_, err = elasticsearch.C().Indices.Create(baseIndexName).Do(ctx)
		if err != nil {
			zap.L().Error("elasticsearch.C().PutIndex()", zap.String("baseIndexName", baseIndexName),
				zap.Error(err))
			return nil, err
		}
	}

	// Create alias "-current" if not already exists
	err = logicalIndex.putAlias(logicalIndexName+"-current", logicalIndexName+"-*", model.Name, ctx)
	if err != nil {
		return nil, err
	}

	// Create alias "-patch" if patchAliasMaxIndices is greater than 0 and alias not already exists
	if model.ElasticsearchOptions.PatchAliasMaxIndices > 0 {
		err = logicalIndex.putAlias(logicalIndexName+"-patch", logicalIndexName+"-*", model.Name, ctx)
		if err != nil {
			return nil, err
		}
	}

	// Create alias "-search" if not already exists (search alias on all active index)
	err = logicalIndex.putAlias(logicalIndexName+"-search", logicalIndexName+"-*", model.Name, ctx)
	if err != nil {
		return nil, err
	}

	// We want to persist the index if he has been created
	if !templateExists {
		err = logicalIndex.persistTechnicalIndex(baseIndexName, now)
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

func (logicalIndex *LogicalIndexCron) putAlias(name, indexPattern, modelName string, ctx context.Context) error {
	aliasExists, err := elasticsearch.C().Indices.ExistsAlias(name).IsSuccess(ctx)
	if err != nil {
		zap.L().Error("IndexExists()", zap.String("aliasName", name), zap.Error(err))
		return err
	}
	if aliasExists {
		return nil
	}

	zap.L().Info("Creating missing alias", zap.String("aliasName", name), zap.String("aliasIndex", indexPattern), zap.String("model", modelName))

	_, err = elasticsearch.C().Indices.PutAlias(indexPattern, name).Do(ctx)
	if err != nil {
		zap.L().Error("elasticsearch.C().PutAlias()", zap.String("aliasName", name), zap.Error(err))
		return err
	}
	return nil
}

func (logicalIndex *LogicalIndexCron) putTemplate(name string, indexPattern string, model modeler.Model) {
	req := elasticsearch.NewPutTemplateRequestV8([]string{indexPattern}, model)

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

func (logicalIndex *LogicalIndexCron) GetCron() *cron.Cron {
	return logicalIndex.Cron
}

// FindIndices search in indices referential every indices between two dates (calculated using current time and depth)
func (logicalIndex *LogicalIndexCron) FindIndices(t time.Time, depthDays int64) ([]string, error) {
	if depthDays == 0 {
		return []string{fmt.Sprintf("%s-%s", logicalIndex.Name, index.All)}, nil
	}

	// default for a day cron
	dateMultiplicationFactor := time.Hour * 24

	// If the cron has entries, adjust the dateMultiplicationFactor based on the first cron entry.
	// This allows the function to support custom cron intervals (not just daily).
	// The factor is calculated as the duration between the start of the year and the next scheduled cron run.
	if len(logicalIndex.Cron.Entries()) != 0 {
		firstEntry := logicalIndex.Cron.Entries()[0]
		startOfTheYear := time.Date(t.Year(), time.January, 1, 0, 0, 0, 0, time.UTC)
		next := firstEntry.Schedule.Next(startOfTheYear)
		dateMultiplicationFactor = next.Sub(startOfTheYear)
	}

	query := "select technical from elasticsearch_indices_v1 where logical = :logical AND creation_date BETWEEN :mindate AND :maxdate"
	params := map[string]interface{}{
		"logical": logicalIndex.Name,
		"mindate": t.Add(-1 * time.Duration(depthDays+1) * dateMultiplicationFactor),
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

// generateNextIndexName generates the next index name based on the current date and existing indices
// Format: logicalIndexName-YYYY-MM-NNNN
func generateNextIndexName(logicalIndexName string, existingIndices []string, now time.Time) string {
	datePrefix := now.Format("2006-01")
	
	// Find the highest sequence number for the current month
	highestSeq := 0
	for _, indexName := range existingIndices {
		// Expected format: logicalIndexName-YYYY-MM-NNNN
		// Minimum length: len(logicalIndexName) + 1 ("-") + 7 ("YYYY-MM") + 1 ("-") + 4 ("NNNN") = len(logicalIndexName) + 13
		expectedLength := len(logicalIndexName) + 13
		if len(indexName) != expectedLength {
			continue
		}
		
		// Check if it starts with logicalIndexName followed by "-"
		if indexName[:len(logicalIndexName)+1] != logicalIndexName+"-" {
			continue
		}
		
		suffix := indexName[len(logicalIndexName)+1:] // Remove "logicalIndexName-"
		// suffix should be exactly "YYYY-MM-NNNN" (12 characters)
		if len(suffix) != 12 {
			continue
		}
		
		// Check if the date prefix matches
		if suffix[:7] != datePrefix {
			continue
		}
		
		// Check if there's a dash separator
		if suffix[7] != '-' {
			continue
		}
		
		// Extract and parse sequence number (last 4 digits)
		seqStr := suffix[8:12]
		seq := 0
		if n, err := fmt.Sscanf(seqStr, "%04d", &seq); err == nil && n == 1 {
			if seq > highestSeq {
				highestSeq = seq
			}
		}
	}
	
	newSeq := highestSeq + 1
	return fmt.Sprintf("%s-%s-%04d", logicalIndexName, datePrefix, newSeq)
}

func (logicalIndex *LogicalIndexCron) rollover() {
	ctx := context.Background()

	// Generate new index name with current date format: logicalIndexName-YYYY-MM-NNNN
	now := time.Now().UTC()
	
	// Find the latest index with the same date prefix to determine the next sequence number
	aliasName := logicalIndex.Name + "-current"
	resAlias, err := elasticsearch.C().Indices.GetAlias().Name(aliasName).Do(ctx)
	
	var newIndexName string
	if err != nil {
		zap.L().Error("GetIndicesByAlias for rollover", zap.Error(err))
		// Fallback to 0001 if we can't determine the current index
		newIndexName = generateNextIndexName(logicalIndex.Name, []string{}, now)
		zap.L().Warn("Could not determine current index, using fallback", zap.String("newIndex", newIndexName))
	} else {
		// Get the current indices
		currentIndices := make([]string, 0)
		for indexName := range resAlias {
			currentIndices = append(currentIndices, indexName)
		}
		sort.Strings(currentIndices)
		
		newIndexName = generateNextIndexName(logicalIndex.Name, currentIndices, now)
		zap.L().Info("Generated new index name for rollover", zap.String("newIndex", newIndexName))
	}

	// Using rollover API to manage alias swap and indices names
	// TODO: Change this dirty abuse of rollover API (triggered every day "max_docx = 0")
	req := rollover.NewRequest()
	// req.Conditions = types.NewRolloverConditions()
	// req.Conditions.MaxAge = "24 h"
	// req.Conditions.MaxDocs = some.Int64(0)

	resRollover, err := elasticsearch.C().Indices.Rollover(logicalIndex.Name + "-current").
		NewIndex(newIndexName).
		Request(req).
		Do(ctx)
	if err != nil {
		zap.L().Error("RollOverV2", zap.Error(err))
		return
	}

	// Roll patch alias
	alias := logicalIndex.Name + "-patch"
	resAlias, err = elasticsearch.C().Indices.GetAlias().Name(alias).Do(ctx)
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

		updateAliasesResponse, err := elasticsearch.C().Indices.UpdateAliases().Request(updateAliasesRequest).Do(ctx)
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
	_, err = elasticsearch.C().Indices.PutAlias(logicalIndex.Name+"-*", logicalIndex.Name+"-search").Do(ctx)
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
		resAlias, err := elasticsearch.C().Indices.GetAlias().Name(logicalIndex.Name + "-search").Do(ctx)
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

			_, err := elasticsearch.C().Indices.Delete(toDelete).Do(ctx)
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

func (logicalIndex *LogicalIndexCron) persistTechnicalIndex(newIndex string, t time.Time) error {
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

func (logicalIndex *LogicalIndexCron) purgeTechnicalIndex(technicalIndex string) error {
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
