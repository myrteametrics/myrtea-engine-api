package coordinator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/myrteametrics/myrtea-sdk/v4/modeler"
	"github.com/myrteametrics/myrtea-sdk/v4/models"

	"github.com/myrteametrics/myrtea-sdk/v4/elasticsearch"
	"github.com/myrteametrics/myrtea-sdk/v4/postgres"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// RollMode is a enumeration of all indices roll mode
type RollMode int

const (
	// Rollover is the native elasticsearch roll mode, based on /_rollover API
	Rollover RollMode = iota + 1
	// Timebased is ...
	Timebased
)

// LogicalIndex abstracts a group a technical elasticsearch indices, which are accessibles with specific aliases
type LogicalIndex struct {
	Initialized bool
	Name        string
	Cron        *cron.Cron
	Executor    *elasticsearch.EsExecutor
	Model       modeler.Model
}

func (logicalIndex *LogicalIndex) initialize() error {

	ctx := context.Background()

	indexPatern := fmt.Sprintf("%s-active-*", logicalIndex.Name)

	exists, err := logicalIndex.Executor.IndexExists(ctx, indexPatern)
	if err != nil {
		return err
	}
	if !exists {

		// Build and put template
		templateName := fmt.Sprintf("template-%s", logicalIndex.Name)
		templateBody := models.NewTemplate(
			[]string{indexPatern},
			logicalIndex.Model.ToElasticsearchMappingProperties(),
			logicalIndex.Model.ElasticsearchOptions.AdvancedSettings,
		)

		b, _ := json.MarshalIndent(templateBody, "", " ")
		zap.L().Debug("template", zap.String("body", string(b)))

		err := logicalIndex.Executor.PutTemplate(ctx, templateName, templateBody)
		if err != nil {
			zap.L().Error("logicalIndex.Executor.PutTemplate()", zap.Error(err))
			return err
		}

		// Put bootstrap index if missing
		err = logicalIndex.Executor.PutIndex(ctx, logicalIndex.Name+"-active-*", logicalIndex.Name+"-active-000001")
		if err != nil {
			zap.L().Error("logicalIndex.Executor.PutIndex()", zap.Error(err))
			return err
		}

		// Adding current active index alias
		err = logicalIndex.Executor.PutAlias(ctx, logicalIndex.Name+"-active-*", logicalIndex.Name+"-current")
		if err != nil {
			zap.L().Error("logicalIndex.Executor.PutAlias()", zap.Error(err))
			return err
		}

		if logicalIndex.Model.ElasticsearchOptions.PatchAliasMaxIndices > 0 {
			// Adding current active index alias
			err = logicalIndex.Executor.PutAlias(ctx, logicalIndex.Name+"-*", logicalIndex.Name+"-patch")
			if err != nil {
				zap.L().Error("logicalIndex.Executor.PutAlias()", zap.Error(err))
				return err
			}
		}

		// Adding search alias on all active index
		err = logicalIndex.Executor.PutAlias(ctx, logicalIndex.Name+"-*", logicalIndex.Name+"-search")
		if err != nil {
			zap.L().Error("logicalIndex.Executor.PutAlias()", zap.Error(err))
			return err
		}

		err = persistTechnicalIndex(logicalIndex.Name, logicalIndex.Name+"-active-000001", time.Now().UTC())
		if err != nil {
			zap.L().Error("Could not persist technical index data", zap.Error(err))
		}
	}

	if logicalIndex.Model.ElasticsearchOptions.Rollmode == "cron" {
		c := cron.New()
		_, err := c.AddFunc(logicalIndex.Model.ElasticsearchOptions.Rollcron, logicalIndex.rollover)
		if err != nil {
			zap.L().Error("Cron add function logicalIndex.updateAliases", zap.Error(err))
			return err
		}
		logicalIndex.Cron = c
		zap.L().Info("Cron started", zap.String("logicalIndex", logicalIndex.Name), zap.String("cron", logicalIndex.Model.ElasticsearchOptions.Rollcron))
	}

	logicalIndex.Initialized = true

	return nil
}

func (logicalIndex *LogicalIndex) rollover() {
	ctx := context.Background()

	// Using rollover API to manage alias swap and indices names
	// TODO: Change this dirty abuse of rollover API (triggered every day "max_docx = 0")
	_, newIndex, err := logicalIndex.Executor.RollOver(ctx, logicalIndex.Name+"-current", "1d", 0)
	if err != nil {
		zap.L().Error("RollOverV2", zap.Error(err), zap.String("index", newIndex))
		return
	}

	// Roll patch alias
	patchIndices, err := logicalIndex.Executor.GetIndicesByAlias(ctx, logicalIndex.Name+"-patch")
	if err != nil {
		zap.L().Error("GetIndicesByAlias", zap.Error(err), zap.String("index", newIndex))
		return
	}
	sort.Strings(patchIndices)

	alias := logicalIndex.Name + "-patch"
	aliasCmd := logicalIndex.Executor.Client.Alias()
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
			err := errors.New("ES API return false acknowledged")
			zap.L().Error("Putting alias", zap.Error(err), zap.String("index", newIndex),
				zap.String("alias", alias))
			return
		}

		zap.L().Info("Patch aliases rollover done", zap.String("alias", alias))
	}

	// Adding search alias on the newly created active index
	// Includes every active + inactive indices
	// TODO: Remove this step when template are reworked (and include this search alias)
	err = logicalIndex.Executor.PutAlias(ctx, logicalIndex.Name+"-*", logicalIndex.Name+"-search")
	if err != nil {
		zap.L().Error("Putting alias", zap.Error(err), zap.String("index", logicalIndex.Name+"-*"),
			zap.String("alias", logicalIndex.Name+"-search"))
		return
	}

	err = persistTechnicalIndex(logicalIndex.Name, newIndex, time.Now().UTC())
	if err != nil {
		zap.L().Error("Could not persist technical index data", zap.Error(err))
	}

	// Purge outdated indices
	if logicalIndex.Model.ElasticsearchOptions.EnablePurge {
		searchIndices, err := logicalIndex.Executor.GetIndicesByAlias(ctx, logicalIndex.Name+"-search")
		if err != nil {
			zap.L().Error("GetIndicesByAlias", zap.Error(err), zap.String("alias", logicalIndex.Name+"-search"))
			return
		}
		sort.Strings(searchIndices)

		if len(searchIndices) > logicalIndex.Model.ElasticsearchOptions.PurgeMaxConcurrentIndices {
			toDelete := searchIndices[0] // oldest indices
			err := logicalIndex.Executor.DeleteIndices(ctx, []string{toDelete})
			if err != nil {
				zap.L().Error("DeleteIndex", zap.Error(err), zap.String("index", searchIndices[0]))
				return
			}
			err = purgeTechnicalIndex(logicalIndex.Name, toDelete)
			if err != nil {
				zap.L().Error("Could not persist technical index data", zap.Error(err))
			}
			zap.L().Info("Index purged", zap.String("name", toDelete))
		}
	}
}

func persistTechnicalIndex(logicalIndex string, newIndex string, t time.Time) error {
	if postgres.DB() == nil {
		return errors.New("Postgresql Client not initialized")
	}

	query := `INSERT INTO elasticsearch_indices_v1 (id, logical, technical, creation_date)
		VALUES (DEFAULT, :logical, :technical, :creation_date);`
	params := map[string]interface{}{
		"logical":       logicalIndex,
		"technical":     newIndex,
		"creation_date": t,
	}

	_, err := postgres.DB().NamedExec(query, params)
	if err != nil {
		return err
	}

	return nil
}

func purgeTechnicalIndex(logicalIndex string, technicalIndex string) error {
	if postgres.DB() == nil {
		return errors.New("Postgresql Client not initialized")
	}

	query := `DELETE FROM elasticsearch_indices_v1 where logical = :logical AND technical = :technical`
	params := map[string]interface{}{
		"logical":   logicalIndex,
		"technical": technicalIndex,
	}

	_, err := postgres.DB().NamedExec(query, params)
	if err != nil {
		return err
	}

	return nil
}
