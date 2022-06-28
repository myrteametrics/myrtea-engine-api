package assistant

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/fact"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/handlers"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/handlers/render"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/reader"
	"github.com/myrteametrics/myrtea-sdk/v4/engine"
	"github.com/myrteametrics/myrtea-sdk/v4/postgres"
	"go.uber.org/zap"
)

func (m *AssistantPlugin) Handler() http.Handler {
	r := chi.NewRouter()

	r.Post("/message", m.SendMessage) // ?time=2019-05-10T12:00:00.000

	return r
}

// Message represents an input raw message with its contexts
type Message struct {
	Sentence      string     `json:"sentence"`
	ContextTokens [][]string `json:"contextTokens"`
}

// SendMessage godoc
// @Summary Sends a message to the myrtea Assistant.
// @Description Sends a message to the myrtea Assistant.
// @Tags Assistant
// @Accept json
// @Produce json
// @Param time query string true "Timestamp"
// @Param sentence body interface{} true "User sentence and context Tokens"
// @Param debug query string false "Enable log debugging"
// @Security Bearer
// @Success 200 "Status OK"
// @Failure 400 "Status Bad Request"
// @Failure 500 "Status Internal Server Error"
// @Router /assistant/message [post]
func (m *AssistantPlugin) SendMessage(w http.ResponseWriter, r *http.Request) {

	// TODO: Refactoring for cleaner code separation
	receiveTs := time.Now().Truncate(1 * time.Millisecond).UTC()

	t, err := handlers.ParseTime(r.URL.Query().Get("time"))
	if err != nil {
		zap.L().Warn("Parse input time", zap.Error(err))
		err := PersistInteractionTrace(receiveTs, t, nil, nil, nil, nil, err)
		if err != nil {
			zap.L().Warn("Persist interaction trace", zap.Error(err))
		}
		render.Error(w, r, render.ErrAPIParsingDateTime, err)
		return
	}

	// debug := false
	// _debug := r.URL.Query().Get("debug")
	// if _debug == "true" {
	// 	debug = true
	// }

	var message Message
	err = json.NewDecoder(r.Body).Decode(&message)
	if err != nil {
		zap.L().Warn("assistant message json decode", zap.Error(err))
		err := PersistInteractionTrace(receiveTs, t, nil, nil, nil, nil, err)
		if err != nil {
			zap.L().Warn("Persist interaction trace", zap.Error(err))
		}
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}

	// f, nlp, err := assistant.C().ProcessMessage(t, message)
	bFact, nlpTokens, err := m.Assistant.SentenceProcess(t.Format("2006-01-02T15:04:05.000Z07:00"), message.Sentence, nil)
	if err != nil {
		zap.L().Warn("NL Process sentence", zap.Error(err))
		render.Error(w, r, render.ErrAPIProcessError, err)
		return
	}

	var f engine.Fact
	err = json.Unmarshal(bFact, &f)
	if err != nil {
		zap.L().Warn("Fact unmarshal", zap.Error(err))
		render.Error(w, r, render.ErrAPIDecodeJSONBody, err)
		return
	}

	for _, dim := range f.Dimensions {
		dim.DateInterval = "1h"
	}

	pf, err := fact.Prepare(&f, -1, -1, t, nil, false)
	if err != nil {
		zap.L().Error("Cannot prepare fact", zap.Error(err))
		render.Error(w, r, render.ErrAPIResourceInvalid, err)
		err := PersistInteractionTrace(receiveTs, t, &message, &nlpTokens, &f, nil, err)
		if err != nil {
			zap.L().Warn("Persist interaction trace", zap.Error(err))
		}
		return
	}

	widgetData, err := fact.Execute(pf)
	if err != nil {
		zap.L().Error("Cannot execute fact", zap.Error(err), zap.Any("prepared-query", pf))
		render.Error(w, r, render.ErrAPIElasticSelectFailed, err)
		err := PersistInteractionTrace(receiveTs, t, &message, &nlpTokens, &f, nil, err)
		if err != nil {
			zap.L().Warn("Persist interaction trace", zap.Error(err))
		}
		return
	}

	err = PersistInteractionTrace(receiveTs, t, &message, &nlpTokens, &f, widgetData.Aggregates, err)
	if err != nil {
		zap.L().Warn("Persist interaction trace", zap.Error(err))
	}

	render.JSON(w, r, &AssistantResponse{
		Result: widgetData,
		Tokens: nlpTokens,
	})
}

// AssistantResponse reflcts the response format of the assistant
type AssistantResponse struct {
	Result *reader.WidgetData `json:"result"`
	Tokens []string           `json:"tokens"`
}

// PersistInteractionTrace store a trace of the assistant interaction in postgresql
func PersistInteractionTrace(receiveTs time.Time, askedTs time.Time, message *Message, nlpTokens *[]string, fact *engine.Fact, result *reader.Item, pipelineErr error) error {

	if postgres.DB() == nil {
		return errors.New("db Client is not initialized")
	}

	params := map[string]interface{}{
		"input_ts":             receiveTs,
		"asked_ts":             askedTs,
		"input_sentence":       nil,
		"input_context_tokens": nil,
		"nlp_tokens":           nil,
		"fact":                 nil,
		"result":               nil,
		"end_ts":               time.Now().Truncate(1 * time.Millisecond).UTC(),
		"success":              false,
		"error":                nil,
	}

	if message != nil {
		params["input_sentence"] = message.Sentence
		if len(message.ContextTokens) > 0 {
			params["input_context_tokens"] = message.ContextTokens[0]
		}
	}

	if pipelineErr == nil {
		params["success"] = true
	} else {
		params["success"] = false
		params["error"] = pipelineErr.Error()
	}

	if factJSON, err := json.Marshal(fact); err == nil {
		params["fact"] = factJSON
	}
	if resultJSON, err := json.Marshal(result); err == nil {
		params["result"] = resultJSON
	}

	query := `
	INSERT INTO interaction_history_v1(id, input_ts, asked_ts, input_sentence, input_context_tokens, nlp_tokens, fact, result, end_ts, success, error)
	VALUES (DEFAULT, :input_ts, :asked_ts, :input_sentence, :input_context_tokens, :nlp_tokens, :fact, :result, :end_ts, :success, :error)`

	res, err := postgres.DB().NamedExec(query, params)
	if err != nil {
		zap.L().Error("Insert interaction history V1", zap.Error(err))
		return err
	}
	i, err := res.RowsAffected()
	if err != nil {
		zap.L().Error("Insert interaction history V1", zap.Error(err))
		return err
	}
	if i != 1 {
		zap.L().Error("Insert interaction history V1", zap.Error(errors.New("no row inserted (or multiple row inserted) instead of 1 row")))
		return errors.New("no row inserted (or multiple row inserted) instead of 1 row")
	}
	return nil
}
