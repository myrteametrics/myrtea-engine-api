package fact

import (
	"context"
	"errors"
	"fmt"
	"time"

	"encoding/json"

	"github.com/jmoiron/sqlx"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/reader"
	"github.com/myrteametrics/myrtea-sdk/v4/builder"
	"github.com/myrteametrics/myrtea-sdk/v4/elasticsearch"
	"github.com/myrteametrics/myrtea-sdk/v4/engine"
	"github.com/myrteametrics/myrtea-sdk/v4/index"
	"github.com/myrteametrics/myrtea-sdk/v4/postgres"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// Prepare enrich a fact by specifying the offset and the number of hits returned
// It also specify the target index based on the fact model
// The
func Prepare(f *engine.Fact, nhit int, offset int, t time.Time, placeholders map[string]string) (*builder.EsSearch, error) {
	if f.AdvancedSource != "" {
		var data map[string]interface{}
		err := json.Unmarshal([]byte(f.AdvancedSource), &data)
		if err != nil {
			return nil, err
		}

		data["size"] = 0
		if nhit != -1 {
			data["size"] = nhit
		}

		data["from"] = 0
		if offset != -1 {
			data["from"] = offset
		}

		b, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}

		strSource := string(b)
		strSource = builder.ContextualizeDatePlaceholders(strSource, t)
		strSource = builder.ContextualizeTimeZonePlaceholders(strSource, t)
		strSource = builder.ContextualizePlaceholders(strSource, placeholders)

		f.AdvancedSource = strSource
	} else {
		f.ContextualizeDimensions(t, placeholders)
		err := f.ContextualizeCondition(t, placeholders)
		if err != nil {
			return nil, err
		}
	}

	esSearch, err := f.ToElasticQuery(t, placeholders)
	if err != nil {
		return nil, err
	}

	// TODO: remove viper.GetString (and use a real configuration object)
	indices, err := FindIndicesWithDynamicDepth(postgres.DB(), viper.GetString("INSTANCE_NAME"), f.Model, t, f.CalculationDepth)
	if err != nil {
		return nil, err
	}
	if len(indices) == 0 {
		zap.L().Info("No indices found, fallback on search-all", zap.String("fact", f.Name), zap.String("model", f.Model), zap.Int64("depth", f.CalculationDepth))
		indices = []string{index.BuildAliasName(viper.GetString("INSTANCE_NAME"), f.Model, index.All)}
		//return nil, fmt.Errorf("No indices found to execute fact %s on depth %d", f.Name, f.CalculationDepth)
	}
	esSearch.Indices = indices

	esSearch.Size = 0
	if nhit != -1 {
		esSearch.Size = nhit
	}

	esSearch.Offset = 0
	if offset != -1 {
		esSearch.Offset = offset
	}

	return esSearch, nil
}

// Execute calculate a fact a specific time and returns the result in a standard format
func Execute(esSearch *builder.EsSearch) (*reader.WidgetData, error) {
	search, err := builder.BuildEsSearch(elasticsearch.C(), esSearch)
	if err != nil {
		return nil, err
	}

	res, err := elasticsearch.C().ExecuteSearch(context.Background(), search)
	if err != nil {
		return nil, err
	}

	widgetData, err := reader.Parse(res)
	if err != nil {
		return nil, err
	}

	return widgetData, nil
}

var depthErrorMarginDays int64 = 1

// FindIndicesWithDynamicDepth search in indices referential every indices between two dates (calculated using current time and depth)
func FindIndicesWithDynamicDepth(dbClient *sqlx.DB, instance string, model string, t time.Time, factCalculationDepthDays int64) ([]string, error) {

	if factCalculationDepthDays == 0 {
		return []string{index.BuildAliasName(instance, model, index.All)}, nil
	}

	if dbClient == nil {
		return nil, errors.New("")
	}

	query := "select technical from elasticsearch_indices_v1 where logical = :logical AND creation_date BETWEEN :mindate AND :maxdate"
	params := map[string]interface{}{
		"logical": fmt.Sprintf("%s-%s", instance, model),
		"mindate": t.Add(-1 * time.Duration(factCalculationDepthDays+depthErrorMarginDays) * 24 * time.Hour),
		"maxdate": t,
	}
	rows, err := dbClient.NamedQuery(query, params)
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
