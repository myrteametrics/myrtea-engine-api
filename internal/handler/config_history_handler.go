package handler

import (
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/config_history"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/security/permissions"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/utils/httputil"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
)

// GetConfigHistories godoc
//
//	@Summary		Get all config history entries
//	@Description	Get all config history entries
//	@Tags			ConfigHistory
//	@Produce		json
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{array}	config_history.ConfigHistory	"list of all config histories"
//	@Failure		500	"internal server error"
//	@Router			/engine/config-histories [get]
func GetConfigHistories(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeConfig, permissions.All, permissions.ActionList)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	histories, err := config_history.R().GetAll()
	if err != nil {
		zap.L().Error("Error getting config histories", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	// Convert map to slice for consistent ordering
	historiesSlice := make([]config_history.ConfigHistory, 0, len(histories))
	for _, h := range histories {
		historiesSlice = append(historiesSlice, h)
	}

	httputil.JSON(w, r, historiesSlice)
}

// GetConfigHistory godoc
//
//	@Summary		Get a config history entry
//	@Description	Get a config history entry by ID
//	@Tags			ConfigHistory
//	@Produce		json
//	@Param			id	path	string	true	"Config History ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	config_history.ConfigHistory	"config history"
//	@Failure		400	"Status Bad Request"
//	@Failure		404	"Status Not Found"
//	@Router			/engine/config-histories/{id} [get]
func GetConfigHistory(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idHistory, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing config history id", zap.String("historyID", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeConfig, permissions.All, permissions.ActionGet)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	history, found, err := config_history.R().Get(idHistory)
	if err != nil {
		zap.L().Error("Cannot get config history", zap.Int64("historyId", idHistory), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("Config history does not exist", zap.Int64("historyId", idHistory))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, err)
		return
	}

	httputil.JSON(w, r, history)
}

// GetConfigHistoriesByType godoc
//
//	@Summary		Get config history entries by type
//	@Description	Get all config history entries of a specific type
//	@Tags			ConfigHistory
//	@Produce		json
//	@Param			type	path	string	true	"History Type"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{array}	config_history.ConfigHistory	"list of config histories by type"
//	@Failure		400	"Status Bad Request"
//	@Router			/engine/config-histories/type/{type} [get]
func GetConfigHistoriesByType(w http.ResponseWriter, r *http.Request) {
	historyType := chi.URLParam(r, "type")

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeConfig, permissions.All, permissions.ActionList)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	histories, err := config_history.R().GetAllByType(historyType)
	if err != nil {
		zap.L().Error("Error getting config histories by type", zap.String("type", historyType), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	// Convert map to slice for consistent ordering
	historiesSlice := make([]config_history.ConfigHistory, 0, len(histories))
	for _, h := range histories {
		historiesSlice = append(historiesSlice, h)
	}

	httputil.JSON(w, r, historiesSlice)
}

// GetConfigHistoriesByUser godoc
//
//	@Summary		Get config history entries by user
//	@Description	Get all config history entries created by a specific user
//	@Tags			ConfigHistory
//	@Produce		json
//	@Param			user	path	string	true	"Username"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{array}	config_history.ConfigHistory	"list of config histories by user"
//	@Failure		400	"Status Bad Request"
//	@Router			/engine/config-histories/user/{user} [get]
func GetConfigHistoriesByUser(w http.ResponseWriter, r *http.Request) {
	user := chi.URLParam(r, "user")

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeConfig, permissions.All, permissions.ActionList)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	histories, err := config_history.R().GetAllByUser(user)
	if err != nil {
		zap.L().Error("Error getting config histories by user", zap.String("user", user), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	// Convert map to slice for consistent ordering
	historiesSlice := make([]config_history.ConfigHistory, 0, len(histories))
	for _, h := range histories {
		historiesSlice = append(historiesSlice, h)
	}

	httputil.JSON(w, r, historiesSlice)
}

// GetConfigHistoriesByInterval godoc
//
//	@Summary		Get config history entries by time interval
//	@Description	Get all config history entries within a specified time interval
//	@Tags			ConfigHistory
//	@Accept			json
//	@Produce		json
//	@Param			interval	body	struct{From string `json:"from"`;To string `json:"to"`}	true	"Time interval (format: 2006-01-02T15:04:05Z)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{array}	config_history.ConfigHistory	"list of config histories in interval"
//	@Failure		400	"Status Bad Request"
//	@Router			/engine/config-histories/interval [post]
func GetConfigHistoriesByInterval(w http.ResponseWriter, r *http.Request) {
	var interval struct {
		From string `json:"from"`
		To   string `json:"to"`
	}

	err := json.NewDecoder(r.Body).Decode(&interval)
	if err != nil {
		zap.L().Warn("Error decoding interval json", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}

	from, err := time.Parse(time.RFC3339, interval.From)
	if err != nil {
		zap.L().Warn("Error parsing from date", zap.String("from", interval.From), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingDateTime, err)
		return
	}

	to, err := time.Parse(time.RFC3339, interval.To)
	if err != nil {
		zap.L().Warn("Error parsing to date", zap.String("to", interval.To), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingDateTime, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeConfig, permissions.All, permissions.ActionList)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	histories, err := config_history.R().GetAllFromInterval(from, to)
	if err != nil {
		zap.L().Error("Error getting config histories by interval", zap.Time("from", from), zap.Time("to", to), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	// Convert map to slice for consistent ordering
	historiesSlice := make([]config_history.ConfigHistory, 0, len(histories))
	for _, h := range histories {
		historiesSlice = append(historiesSlice, h)
	}

	httputil.JSON(w, r, historiesSlice)
}

// CreateConfigHistory godoc
//
//	@Summary		Create a new config history entry
//	@Description	Create a new config history entry
//	@Tags			ConfigHistory
//	@Accept			json
//	@Produce		json
//	@Param			history	body	struct{Commentary string `json:"commentary"`;Type string `json:"type"`;User string `json:"user"`}	true	"Config History (json)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	config_history.ConfigHistory	"created config history"
//	@Failure		400	"Status Bad Request"
//	@Failure		500	"internal server error"
//	@Router			/engine/config-histories [post]
func CreateConfigHistory(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeConfig, permissions.All, permissions.ActionCreate)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var historyInput struct {
		Commentary string `json:"commentary"`
		Type       string `json:"type"`
		User       string `json:"user"`
	}

	err := json.NewDecoder(r.Body).Decode(&historyInput)
	if err != nil {
		zap.L().Warn("Config history json decoding", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}

	// Create a new history entry with auto-generated ID
	newHistory := config_history.NewConfigHistory(historyInput.Commentary, historyInput.Type, historyInput.User)

	if ok, err := newHistory.IsValid(); !ok {
		zap.L().Warn("Config history is not valid", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIResourceInvalid, err)
		return
	}

	id, err := config_history.R().Create(newHistory)
	if err != nil {
		zap.L().Error("Error creating config history", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBInsertFailed, err)
		return
	}

	createdHistory, found, err := config_history.R().Get(id)
	if err != nil {
		zap.L().Error("Error getting created config history", zap.Int64("id", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("Created config history not found", zap.Int64("id", id))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFoundAfterInsert, errors.New("created config history not found"))
		return
	}

	httputil.JSON(w, r, createdHistory)
}

// DeleteConfigHistory godoc
//
//	@Summary		Delete a config history entry
//	@Description	Delete a config history entry by ID
//	@Tags			ConfigHistory
//	@Produce		json
//	@Param			id	path	string	true	"Config History ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	"Status OK"
//	@Failure		400	"Status Bad Request"
//	@Router			/engine/config-histories/{id} [delete]
func DeleteConfigHistory(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idHistory, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing config history id", zap.String("historyID", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeConfig, permissions.All, permissions.ActionDelete)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	err = config_history.R().Delete(idHistory)
	if err != nil {
		zap.L().Error("Error deleting config history", zap.Int64("id", idHistory), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBDeleteFailed, err)
		return
	}

	httputil.OK(w, r)
}

// DeleteOldestConfigHistory godoc
//
//	@Summary		Delete the oldest config history entry
//	@Description	Delete the oldest config history entry (the one with the lowest ID)
//	@Tags			ConfigHistory
//	@Produce		json
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	"Status OK"
//	@Failure		500	"internal server error"
//	@Router			/engine/config-histories/oldest [delete]
func DeleteOldestConfigHistory(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeConfig, permissions.All, permissions.ActionDelete)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	err := config_history.R().DeleteOldest()
	if err != nil {
		zap.L().Error("Error deleting oldest config history", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBDeleteFailed, err)
		return
	}

	httputil.OK(w, r)
}
