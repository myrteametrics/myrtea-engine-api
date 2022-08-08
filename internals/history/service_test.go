package history

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/myrteametrics/myrtea-engine-api/v4/internals/fact"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/models"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/reader"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/scheduler"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/situation"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/tests"
	"github.com/myrteametrics/myrtea-sdk/v4/expression"
	"go.uber.org/zap"
)

// builder.newStatement().Update("situation_history_v4").
// 	Where("id = ?", history.ID).
// 	Set("metadatas", history.Metadatas)
// Set("last_update", time.Now())

// builder.newStatement().Update("situation_history_v4").
// 	Where("id = ?", history.ID).
// 	Set("situation_id", history.SituationID).
// 	Set("situation_instance_id", history.SituationInstanceID).
// 	Set("ts", history.Ts).
// 	Set("parameters", history.Parameters).
// 	Set("expression_facts", history.ExpressionFacts).
// 	Set("metadatas", history.Metadatas)
// Set("last_update", time.Now())

func CalculateFact(ti time.Time) reader.Item {
	return reader.Item{Aggs: map[string]*reader.ItemAgg{"doc_count": {Value: 100}}}
}

func (service HistoryService) ExtractSituationData(situationID int64, situationInstanceID int64) (situation.Situation, string, map[string]interface{}, error) {
	parameters := make(map[string]interface{})

	s, found, err := situation.R().Get(situationID)
	if err != nil {
		return situation.Situation{}, "", make(map[string]interface{}), err
	}
	if !found {
		return situation.Situation{}, "", make(map[string]interface{}), errors.New("situation not found")
	}
	for k, v := range s.Parameters {
		parameters[k] = v
	}

	situationInstanceName := ""
	if s.IsTemplate {
		si, found, err := situation.R().GetTemplateInstance(situationInstanceID)
		if err != nil {
			return situation.Situation{}, "", make(map[string]interface{}), err
		}
		if !found {
			return situation.Situation{}, "", make(map[string]interface{}), errors.New("situation instance not found")
		}
		situationInstanceName = si.Name
		for k, v := range si.Parameters {
			parameters[k] = v
		}
	}

	return s, situationInstanceName, parameters, nil
}

// Flatten situation data (old and new facts + parameters)
func (service HistoryService) ExtractFactData(historyFactNew HistoryFactsV4, existingFactIDs []int64) ([]HistoryFactsV4, map[string]interface{}, error) {
	historySituationFlattenData := make(map[string]interface{})
	historyFactsNew := make([]HistoryFactsV4, 0)
	for _, factId := range existingFactIDs {
		if factId == historyFactNew.FactID {
			historyFactNewData, err := historyFactNew.Result.ToAbstractMap()
			if err != nil {
				// err
			}

			historySituationFlattenData[historyFactNew.FactName] = historyFactNewData
			historyFactsNew = append(historyFactsNew, historyFactNew)

		} else {
			query := service.historyFactsQuerier.Builder.GetHistoryFactLast(factId)
			historyFactLast, err := service.historyFactsQuerier.QueryOne(query)
			if err != nil {
				// err
			}

			historyFactLastData, err := historyFactLast.Result.ToAbstractMap()
			if err != nil {
				// err
			}

			historySituationFlattenData[historyFactLast.FactName] = historyFactLastData
			historyFactsNew = append(historyFactsNew, historyFactLast)
		}
	}
	return historyFactsNew, historySituationFlattenData, nil
}

func EvaluateExpressionFacts(expressionFacts []situation.ExpressionFact, historySituationFlattenData map[string]interface{}) map[string]interface{} {
	expressionFactsEvaluated := make(map[string]interface{})
	for _, expressionFact := range expressionFacts {
		result, err := expression.Process(expression.LangEval, expressionFact.Expression, historySituationFlattenData)
		if err != nil {
			continue
		}
		if expression.IsInvalidNumber(result) {
			continue
		}

		historySituationFlattenData[expressionFact.Name] = result
		expressionFactsEvaluated[expressionFact.Name] = result
	}
	return expressionFactsEvaluated
}

func (historyService HistoryService) CalculateAndStoreFacts(ti time.Time, factIDs []int64) ([]HistoryFactsV4, error) {

	historyFactsNew := make([]HistoryFactsV4, 0)

	for _, factID := range factIDs {
		f, found, err := fact.R().Get(factID)
		if err != nil {
			zap.L().Error("Error Getting the Fact, skipping fact calculation...", zap.Int64("factID", factID))
			continue
		}
		if !found {
			zap.L().Warn("Fact does not exists, skipping fact calculation...", zap.Int64("factID", factID))
			continue
		}

		// Get Situations Linked to fact => For each calculate + persist
		situationHistoryRecordV1, err := scheduler.GetFactSituations(f, ti)
		if err != nil {
			continue
		}
		if len(situationHistoryRecordV1) == 0 {
			zap.L().Info("No situation within valid calendar period for the Fact, skipping fact calculation...", zap.Int64("factID", factID))
			// S().RemoveRunningJob(job.ScheduleID)
			// return
		}

		// Calculate Fact
		factValue := CalculateFact(ti)

		// Insert fact history
		historyFactNew := HistoryFactsV4{
			// ID: -1,
			//FactName: "",
			SituationID:         1,
			SituationInstanceID: 1,
			FactID:              1,
			Ts:                  ti,
			Result:              factValue,
		}
		historyFactNew.ID, err = historyService.historyFactsQuerier.Insert(historyFactNew)
		if err != nil {
			// err
		}

		historyFactsNew = append(historyFactsNew, historyFactNew)
	}
	return historyFactsNew, nil
}

func TestQuery5(t *testing.T) {
	t.Fail()
	db := tests.DBClient(t)
	historyService := New(db)
	situation.ReplaceGlobals(situation.NewPostgresRepository(tests.DBClient(t)))

	var err error
	ti := time.Now()

	factIDs := []int64{1, 2, 14, 15}

	// Calculate and persist facts => Retrieve all HistoryFactV4 with ID
	historyFactsNew, err := historyService.CalculateAndStoreFacts(ti, factIDs)
	if err != nil {
		t.Error(err)
	}

	historyFactNew := historyFactsNew[0]

	// Flatten parameters from situation definition + situation instance definition
	s, situationInstanceName, parameters, err := historyService.ExtractSituationData(historyFactNew.SituationID, historyFactNew.SituationInstanceID)
	if err != nil {
		t.Error(err)
	}
	t.Log("parameters", parameters)

	historyFactsNew, historySituationFlattenData, err := historyService.ExtractFactData(historyFactNew, s.Facts)
	if err != nil {
		t.Error(err)
	}
	for key, value := range parameters {
		historySituationFlattenData[key] = value
	}
	for key, value := range expression.GetDateKeywords(ti) {
		historySituationFlattenData[key] = value
	}
	t.Log("flatten data", historySituationFlattenData)

	// Evaluate expression facts
	expressionFacts := EvaluateExpressionFacts(s.ExpressionFacts, historySituationFlattenData)
	t.Log("expressionfacts", expressionFacts)

	// Build and insert HistorySituationV4
	historySituationNew := HistorySituationsV4{
		// ID:                    -1,
		SituationName:         s.Name,
		SituationInstanceName: situationInstanceName,
		SituationID:           historyFactNew.SituationID,
		SituationInstanceID:   historyFactNew.SituationInstanceID,
		Ts:                    historyFactNew.Ts,
		Parameters:            parameters,
		ExpressionFacts:       expressionFacts,
		Metadatas:             make([]models.MetaData, 0),
	}
	historySituationNew.ID, err = historyService.historySituationsQuerier.Insert(historySituationNew)
	if err != nil {
		t.Error(err)
	}
	t.Log("historySituationNew", historySituationNew)

	// Build and insert HistorySituationFactsV4
	historySituationFactNew := make([]HistorySituationFactsV4, 0)
	for _, historyFactNew := range historyFactsNew {
		historySituationFactNew = append(historySituationFactNew, HistorySituationFactsV4{ // Replace entry for existing factID with new HistorySituationFactsV4{}
			HistorySituationID: historySituationNew.ID,
			HistoryFactID:      historyFactNew.ID,
			FactID:             historyFactNew.FactID,
		})
	}

	query := historyService.historySituationFactsQuerier.Builder.InsertBulk(historySituationFactNew)
	err = historyService.historySituationFactsQuerier.execute(query)
	if err != nil {
		t.Error(err)
	}
	t.Log("historySituationFactNew", historySituationFactNew)
}

// func TestQuery4(t *testing.T) {
// 	t.Fail()
// 	db := tests.DBClient(t)
// 	historyService := New(db)
// 	situation.ReplaceGlobals(situation.NewPostgresRepository(tests.DBClient(t)))

// 	/*
// 		Receive facts to calculate
// 		=> 1, 2, 3

// 		Search associated situation from settings
// 		=> 1, 5

// 		Check if template (and search for instances) or if simple situation

// 		Calculate the fact from source (ES)
// 		=> 1 = doc_count/value
// 		=> 2 = doc_count/value
// 		=> 3 = doc_count/value

// 		!!!!!! Persist the fact in history table

// 		!!!!!! CAN ALSO BE AN UPDATE ? => Retroactive calculation
// 	*/

// 	/*
// 		Build a list of skeleton SituationHistoryRecord from the calculated facts
// 		(situationID + situationInstanceID + Flatten parameters (situation+instance) )
// 		!! + FactIDS (Combo FactID + TS)

// 		for each situation
// 		get all factIDS inside situation (from settings)

// 		Search for the last-closest SituationHistoryRecord ??????
// 		lastHistoryRecord, err := situation.GetFromHistory(record.ID, record.TS, record.TemplateInstanceID, true)

// 		Merge old record with new calculated fact

// 		Process ExpressionFacts

// 		Persist situation
// 	*/

// 	ti := time.Now()
// 	historyFactNew := HistoryFactsV4{
// 		// ID: -1,
// 		//FactName: "",
// 		SituationID:         1,
// 		SituationInstanceID: 1,
// 		FactID:              1,
// 		Ts:                  ti,
// 		Result:              reader.Item{Aggs: map[string]*reader.ItemAgg{"doc_count": {Value: 100}}},
// 	}

// 	var err error
// 	historyFactNew.ID, err = historyService.historyFactsQuerier.Insert(historyFactNew)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	t.Log(historyFactNew.ID)

// 	// Flatten parameters from situation definition + situation instance definition
// 	situationInstanceName := ""
// 	parameters := make(map[string]interface{})
// 	s, _, _ := situation.R().Get(historyFactNew.SituationID) // .SituationInstanceID
// 	for k, v := range s.Parameters {
// 		parameters[k] = v
// 	}
// 	if s.IsTemplate {
// 		si, _, _ := situation.R().GetTemplateInstance(historyFactNew.SituationInstanceID)
// 		situationInstanceName = si.Name
// 		for k, v := range si.Parameters {
// 			parameters[k] = v
// 		}
// 	}
// 	t.Log("parameters", parameters)

// 	// TODO: Might be faster to just get last fact value ? What if no situation history is found ?
// 	// Flatten facts data (old facts + new fact)
// 	// Get last saved situation
// 	historySituationLasts, err := historyService.historySituationsQuerier.GetHistorySituationsIdsLast(GetHistorySituationsOptions{
// 		SituationID:         historyFactNew.SituationID,
// 		SituationInstanceID: historyFactNew.SituationInstanceID,
// 		ToTS:                historyFactNew.Ts,
// 	})
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	t.Log("last situations", historySituationLasts)

// 	historySituationIDs := make([]int64, 0)
// 	for _, historySituation := range historySituationLasts {
// 		historySituationIDs = append(historySituationIDs, historySituation.ID)
// 	}

// 	// Get last saved facts for the old situation
// 	historyFactLasts, historySituationFactLasts, err := historyService.historyFactsQuerier.GetHistoryFactsFromSituation(historyService.historySituationFactsQuerier, historySituationIDs)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	t.Log("last facts", historyFactLasts)

// 	// Flatten situation data (old and new facts + parameters)
// 	historySituationFlattenData := make(map[string]interface{})
// 	for _, historyFactLast := range historyFactLasts {
// 		historyFactData, err := historyFactLast.Result.ToAbstractMap()
// 		if err != nil {
// 			t.Error(err)
// 		}
// 		historySituationFlattenData[historyFactLast.FactName] = historyFactData
// 	}
// 	historySituationFlattenData[historyFactNew.FactName] = historyFactNew // override datas with new fact data
// 	for key, value := range parameters {
// 		historySituationFlattenData[key] = value
// 	}
// 	for key, value := range expression.GetDateKeywords(ti) {
// 		historySituationFlattenData[key] = value
// 	}
// 	t.Log("flatten data", historySituationFlattenData)

// 	// Evaluate expression facts
// 	expressionFacts := make(map[string]interface{})
// 	for _, expressionFact := range s.ExpressionFacts {
// 		result, err := expression.Process(expression.LangEval, expressionFact.Expression, historySituationFlattenData)
// 		if err != nil {
// 			t.Error(err)
// 			continue
// 		}
// 		if expression.IsInvalidNumber(result) {
// 			continue
// 		}

// 		historySituationFlattenData[expressionFact.Name] = result // Used for chaining expression facts
// 		expressionFacts[expressionFact.Name] = result
// 	}
// 	t.Log("expressionfacts", expressionFacts)

// 	historySituationNew := HistorySituationsV4{
// 		ID:                    -1,
// 		SituationName:         s.Name,
// 		SituationInstanceName: situationInstanceName,
// 		SituationID:           historyFactNew.SituationID,
// 		SituationInstanceID:   historyFactNew.SituationInstanceID,
// 		Ts:                    historyFactNew.Ts,
// 		Parameters:            parameters,
// 		ExpressionFacts:       expressionFacts,
// 		Metadatas:             make([]models.MetaData, 0),
// 	}
// 	historySituationNew.ID, err = historyService.historySituationsQuerier.Insert(historySituationNew)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	t.Log("historySituationNew", historySituationNew)

// 	historySituationFactNewMap := make(map[int64]HistorySituationFactsV4)
// 	for _, historySituationFact := range historySituationFactLasts {
// 		historySituationFactNewMap[historySituationFact.FactID] = HistorySituationFactsV4{
// 			HistorySituationID: historySituationNew.ID, // Replace HistorySituationID with new inserted situation
// 			HistoryFactID:      historySituationFact.HistoryFactID,
// 			FactID:             historySituationFact.FactID,
// 		}
// 	}
// 	historySituationFactNewMap[historyFactNew.FactID] = HistorySituationFactsV4{ // Replace entry for existing factID with new HistorySituationFactsV4{}
// 		HistorySituationID: historySituationNew.ID,
// 		HistoryFactID:      historyFactNew.ID,
// 		FactID:             historyFactNew.FactID,
// 	}

// 	historySituationFactNew := make([]HistorySituationFactsV4, 0, len(historySituationFactNewMap))
// 	for _, val := range historySituationFactNewMap {
// 		historySituationFactNew = append(historySituationFactNew, val)
// 	}

// 	query := historyService.historySituationFactsQuerier.Builder.InsertBulk(historySituationFactNew)
// 	err = historyService.historySituationFactsQuerier.execute(query)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	t.Log("historySituationFactNew", historySituationFactNew)

// 	// override historySituationFacts all entry with new situationHistoryID
// 	// override historySituationFacts one entry with new fact

// 	// insert situation
// 	// insert historySituationFacts (rewrited)

// }

func TestQuery3(t *testing.T) {
	t.Fail()
	db := tests.DBClient(t)
	historyService := New(db)

	options := GetHistorySituationsOptions{
		SituationID:         4,
		SituationInstanceID: -1,
		FromTS:              time.Date(2022, time.July, 1, 0, 0, 0, 0, time.UTC),
		ToTS:                time.Time{},
	}
	interval := "day"

	// Fetch situations history
	historySituations, err := historyService.historySituationsQuerier.GetHistorySituationsIdsByStandardInterval(options, interval)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	// Fetch facts history
	historySituationsIds := make([]int64, 0)
	for _, item := range historySituations {
		historySituationsIds = append(historySituationsIds, item.ID)
	}

	historyFacts, historySituationFacts, err := historyService.historyFactsQuerier.GetHistoryFactsFromSituation(historyService.historySituationFactsQuerier, historySituationsIds)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	// Extract results
	result := historyService.ExtractData(historySituations, historySituationFacts, historyFacts)
	b, _ := json.Marshal(result)
	t.Log(string(b))
}
