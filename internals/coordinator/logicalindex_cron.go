package coordinator

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/myrteametrics/myrtea-sdk/v4/elasticsearchv6"
	"github.com/myrteametrics/myrtea-sdk/v4/index"
	"github.com/myrteametrics/myrtea-sdk/v4/modeler"
	"github.com/myrteametrics/myrtea-sdk/v4/models"

	"github.com/myrteametrics/myrtea-sdk/v4/postgres"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// LogicalIndex abstracts a group a technical elasticsearchv6 indices, which are accessibles with specific aliases
type LogicalIndexCron struct {
	Initialized bool
	Name        string
	Cron        *cron.Cron
	Model       modeler.Model
	mu          sync.RWMutex
}

func NewLogicalIndexCron(instanceName string, model modeler.Model) (*LogicalIndexCron, error) {

	logicalIndexName := fmt.Sprintf("%s-%s", instanceName, model.Name)

	zap.L().Info("Initialize logicalIndex (LogicalIndexCron)", zap.String("name", logicalIndexName), zap.String("model", model.Name), zap.Any("options", model.ElasticsearchOptions))

	if model.ElasticsearchOptions.Rollmode != "cron" {
		return nil, errors.New("invalid rollmode for this logicalIndex type")
	}

	logicalIndex := &LogicalIndexCron{
		Initialized: false,
		Name:        logicalIndexName,
		Cron:        nil,
		Model:       model,
	}

	ctx := context.Background()
	indexPatern := fmt.Sprintf("%s-active-*", logicalIndexName)
	exists, err := elasticsearchv6.C().IndexExists(ctx, indexPatern)
	if err != nil {
		return nil, err
	}
	if !exists {
		// Build and put template
		templateName := fmt.Sprintf("template-%s", logicalIndexName)
		templateBody := models.NewTemplateV6(
			[]string{indexPatern},
			model.ToElasticsearchMappingProperties(),
			model.ElasticsearchOptions.AdvancedSettings,
		)

		err := elasticsearchv6.C().PutTemplate(ctx, templateName, templateBody)
		if err != nil {
			zap.L().Error("elasticsearchv6.C().PutTemplate()", zap.Error(err))
			return nil, err
		}

		// Put bootstrap index if missing
		err = elasticsearchv6.C().PutIndex(ctx, logicalIndexName+"-active-*", logicalIndexName+"-active-000001")
		if err != nil {
			zap.L().Error("elasticsearchv6.C().PutIndex()", zap.Error(err))
			return nil, err
		}

		// Adding current active index alias
		err = elasticsearchv6.C().PutAlias(ctx, logicalIndexName+"-active-*", logicalIndexName+"-current")
		if err != nil {
			zap.L().Error("elasticsearchv6.C().PutAlias()", zap.Error(err))
			return nil, err
		}

		if model.ElasticsearchOptions.PatchAliasMaxIndices > 0 {
			// Adding current active index alias
			err = elasticsearchv6.C().PutAlias(ctx, logicalIndexName+"-*", logicalIndexName+"-patch")
			if err != nil {
				zap.L().Error("elasticsearchv6.C().PutAlias()", zap.Error(err))
				return nil, err
			}
		}

		// Adding search alias on all active index
		err = elasticsearchv6.C().PutAlias(ctx, logicalIndexName+"-*", logicalIndexName+"-search")
		if err != nil {
			zap.L().Error("elasticsearchv6.C().PutAlias()", zap.Error(err))
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

func (logicalIndex *LogicalIndexCron) GetCron() *cron.Cron {
	return logicalIndex.Cron
}

// FindIndicesWithDynamicDepth search in indices referential every indices between two dates (calculated using current time and depth)
func (logicalIndex *LogicalIndexCron) FindIndices(t time.Time, depthDays int64) ([]string, error) {
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

func (logicalIndex *LogicalIndexCron) rollover() {
	ctx := context.Background()

	// Using rollover API to manage alias swap and indices names
	// TODO: Change this dirty abuse of rollover API (triggered every day "max_docx = 0")
	_, newIndex, err := elasticsearchv6.C().RollOver(ctx, logicalIndex.Name+"-current", "1d", 0)
	if err != nil {
		zap.L().Error("RollOverV2", zap.Error(err), zap.String("index", newIndex))
		return
	}

	// Roll patch alias
	patchIndices, err := elasticsearchv6.C().GetIndicesByAlias(ctx, logicalIndex.Name+"-patch")
	if err != nil {
		zap.L().Error("GetIndicesByAlias", zap.Error(err), zap.String("index", newIndex))
		return
	}
	sort.Strings(patchIndices)

	alias := logicalIndex.Name + "-patch"
	aliasCmd := elasticsearchv6.C().Client.Alias()
	if len(patchIndices) <= logicalIndex.Model.ElasticsearchOptions.PatchAliasMaxIndices {
		if len(patchIndices) >= logicalIndex.Model.ElasticsearchOptions.PatchAliasMaxIndices {
			aliasCmd = aliasCmd.Remove(patchIndices[0], alias)
		}
		aliasCmd = aliasCmd.Add(newIndex, alias)
		aliasResult, err := aliasCmd.Do(context.Background())

		if err != nil {
			zap.L().Error("Putting alias", zap.Error(err), zap.String("index", newIndex),
				zap.String("alias", alias))
			return
		}
		if !aliasResult.Acknowledged {
			err := errors.New("es API return false acknowledged")
			zap.L().Error("Putting alias", zap.Error(err), zap.String("index", newIndex),
				zap.String("alias", alias))
			return
		}

		zap.L().Info("Patch aliases rollover done", zap.String("alias", alias))
	}

	// Adding search alias on the newly created active index
	// Includes every active + inactive indices
	// TODO: Remove this step when template are reworked (and include this search alias)
	err = elasticsearchv6.C().PutAlias(ctx, logicalIndex.Name+"-*", logicalIndex.Name+"-search")
	if err != nil {
		zap.L().Error("Putting alias", zap.Error(err), zap.String("index", logicalIndex.Name+"-*"),
			zap.String("alias", logicalIndex.Name+"-search"))
		return
	}

	err = logicalIndex.persistTechnicalIndex(newIndex, time.Now().UTC())
	if err != nil {
		zap.L().Error("Could not persist technical index data", zap.Error(err))
	}

	// Purge outdated indices
	if logicalIndex.Model.ElasticsearchOptions.EnablePurge {
		searchIndices, err := elasticsearchv6.C().GetIndicesByAlias(ctx, logicalIndex.Name+"-search")
		if err != nil {
			zap.L().Error("GetIndicesByAlias", zap.Error(err), zap.String("alias", logicalIndex.Name+"-search"))
			return
		}
		sort.Strings(searchIndices)

		if len(searchIndices) > logicalIndex.Model.ElasticsearchOptions.PurgeMaxConcurrentIndices {
			toDelete := searchIndices[0] // oldest indices
			err := elasticsearchv6.C().DeleteIndices(ctx, []string{toDelete})
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
