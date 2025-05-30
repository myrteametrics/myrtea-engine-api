package scheduler

import (
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	calendar2 "github.com/myrteametrics/myrtea-engine-api/v5/pkg/calendar"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/reader"
	situation2 "github.com/myrteametrics/myrtea-engine-api/v5/pkg/situation"
	elasticsearchsdk "github.com/myrteametrics/myrtea-sdk/v5/elasticsearch"
	"log"
	"testing"
	"time"

	"github.com/myrteametrics/myrtea-engine-api/v5/internal/fact"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/rule"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/tests"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/history"
	"github.com/myrteametrics/myrtea-sdk/v5/helpers"
	"github.com/myrteametrics/myrtea-sdk/v5/postgres"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func loadConfiguration() {

	if viper.GetString("INSTANCE_NAME") == "" {

		allowedConfigKey := [][]helpers.ConfigKey{
			{{Type: helpers.StringFlag, Name: "INSTANCE_NAME", DefaultValue: "myrtea", Description: "Myrtea instance name"}},
		}
		configName, configPath, envPrefix := "engine-api", "config", "MYRTEA"
		helpers.InitializeConfig(allowedConfigKey, configName, configPath, envPrefix)
	}
}

func testInsertSituationHistory(t *testing.T, situationID int64, instanceID int64) []int64 {
	parameters := map[string]interface{}{
		"heure_deadline":              "12",
		"label":                       "Export",
		"seuil_alerte_apres_deadline": "0.6",
		"seuil_alerte_avant_deadline": "0.6",
		"seuil_status":                "0.7",
		"code_pays":                   "US",
		"heure_aprem":                 "14h",
		"heure_matin":                 "11h",
		"seuil_critique":              "0.8",
		"seuil_warning":               "0.9",
	}
	ids1, _ := history.S().HistorySituationsQuerier.Insert(history.HistorySituationsV4{SituationID: situationID, SituationInstanceID: instanceID, Ts: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC), Parameters: parameters})
	ids2, _ := history.S().HistorySituationsQuerier.Insert(history.HistorySituationsV4{SituationID: situationID, SituationInstanceID: instanceID, Ts: time.Date(2023, 1, 1, 18, 0, 0, 0, time.UTC), Parameters: parameters})
	ids3, _ := history.S().HistorySituationsQuerier.Insert(history.HistorySituationsV4{SituationID: situationID, SituationInstanceID: instanceID, Ts: time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC), Parameters: parameters})
	ids4, _ := history.S().HistorySituationsQuerier.Insert(history.HistorySituationsV4{SituationID: situationID, SituationInstanceID: instanceID, Ts: time.Date(2023, 1, 2, 18, 0, 0, 0, time.UTC), Parameters: parameters})
	return []int64{ids1, ids2, ids3, ids4}
}
func testInsertFactHistory(t *testing.T, situationID int64, instanceID int64, factIDs []int64, situationHistoryIDs []int64) {
	res := reader.Item{Aggs: map[string]*reader.ItemAgg{"doc_count": {Value: 99}}}

	for _, factID := range factIDs {
		idf1, _ := history.S().HistoryFactsQuerier.Insert(history.HistoryFactsV4{FactID: factID, SituationID: situationID, SituationInstanceID: instanceID, Ts: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC), Result: res})
		idf2, _ := history.S().HistoryFactsQuerier.Insert(history.HistoryFactsV4{FactID: factID, SituationID: situationID, SituationInstanceID: instanceID, Ts: time.Date(2023, 1, 1, 18, 0, 0, 0, time.UTC), Result: res})
		idf3, _ := history.S().HistoryFactsQuerier.Insert(history.HistoryFactsV4{FactID: factID, SituationID: situationID, SituationInstanceID: instanceID, Ts: time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC), Result: res})
		idf4, _ := history.S().HistoryFactsQuerier.Insert(history.HistoryFactsV4{FactID: factID, SituationID: situationID, SituationInstanceID: instanceID, Ts: time.Date(2023, 1, 2, 18, 0, 0, 0, time.UTC), Result: res})

		_ = history.S().HistorySituationFactsQuerier.Execute(history.S().HistorySituationFactsQuerier.Builder.InsertBulk([]history.HistorySituationFactsV4{
			{HistorySituationID: situationHistoryIDs[0], HistoryFactID: idf1, FactID: factID},
			{HistorySituationID: situationHistoryIDs[1], HistoryFactID: idf2, FactID: factID},
			{HistorySituationID: situationHistoryIDs[2], HistoryFactID: idf3, FactID: factID},
			{HistorySituationID: situationHistoryIDs[3], HistoryFactID: idf4, FactID: factID},
		}))
	}
}

func TestFactRecalculationJobRun(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping DB test in short mode")
	}
	zapCfg := zap.NewDevelopmentConfig()
	zapCfg.Level.SetLevel(zap.InfoLevel)
	logger, err := zapCfg.Build(zap.AddStacktrace(zap.ErrorLevel))
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	defer logger.Sync()
	zap.ReplaceGlobals(logger)

	loadConfiguration()
	elasticsearchsdk.ReplaceGlobals(elasticsearch.Config{
		Addresses: []string{"http://localhost:9200"},
	})

	db := tests.DBClient(t)
	postgres.ReplaceGlobals(db)
	history.ReplaceGlobals(history.New(db))
	fact.ReplaceGlobals(fact.NewPostgresRepository(db))
	situation2.ReplaceGlobals(situation2.NewPostgresRepository(db))
	rule.ReplaceGlobals(rule.NewPostgresRepository(db))
	calendar2.ReplaceGlobals(calendar2.NewPostgresRepository(db))
	calendar2.Init()
	ReplaceGlobalRepository(NewPostgresRepository(db))
	ReplaceGlobals(NewScheduler())
	S().Init()

	db.Exec("truncate situation_fact_history_v5 cascade")
	db.Exec("truncate fact_history_v5 cascade")
	db.Exec("truncate situation_history_v5 cascade")
	// db.Exec("SELECT setval(pg_get_serial_sequence('situation_fact_history_v5', 'id'), coalesce(max(id),0) + 1, false) FROM situation_fact_history_v5")
	db.Exec("SELECT setval(pg_get_serial_sequence('fact_history_v5', 'id'), coalesce(max(id),0) + 1, false) FROM fact_history_v5")
	db.Exec("SELECT setval(pg_get_serial_sequence('situation_history_v5', 'id'), coalesce(max(id),0) + 1, false) FROM situation_history_v5")

	// non-template
	historySituation17 := testInsertSituationHistory(t, 17, 0)
	testInsertFactHistory(t, 0, 0, []int64{185, 187}, historySituation17)

	// template
	historySituation121 := testInsertSituationHistory(t, 12, 1)
	testInsertFactHistory(t, 12, 1, []int64{152, 153, 154, 158, 159, 160, 172, 214}, historySituation121)

	historySituation122 := testInsertSituationHistory(t, 12, 2)
	testInsertFactHistory(t, 12, 2, []int64{152, 153, 154, 158, 159, 160, 172, 214}, historySituation122)

	historySituation123 := testInsertSituationHistory(t, 12, 3)
	testInsertFactHistory(t, 12, 3, []int64{152, 153, 154, 158, 159, 160, 172, 214}, historySituation123)

	historySituation124 := testInsertSituationHistory(t, 12, 4)
	testInsertFactHistory(t, 12, 4, []int64{152, 153, 154, 158, 159, 160, 172, 214}, historySituation124)

	job := FactRecalculationJob{
		FactIds:        []int64{152, 185},
		From:           fmt.Sprintf(`"%s"`, time.Date(2023, 1, 4, 0, 0, 0, 0, time.UTC).Add(-15*24*time.Hour).Format(timeLayout)),
		To:             fmt.Sprintf(`"%s"`, time.Date(2023, 1, 4, 0, 0, 0, 0, time.UTC).Format(timeLayout)),
		LastDailyValue: true,
	}
	job.Run()

	t.Fail()
}
