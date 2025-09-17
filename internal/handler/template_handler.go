package handler

import (
	"encoding/json"
	"errors"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/email/template"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/security/permissions"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/utils/httputil"
	"net/http"
	"sort"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// GetTemplates godoc
//
//	@Summary		Get all email templates
//	@Description	Get all email templates
//	@Tags			Templates
//	@Produce		json
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{array}	template.Template	"list of email templates"
//	@Failure		500	"internal server error"
//	@Router			/engine/templates [get]
func GetTemplates(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeTemplate, permissions.All, permissions.ActionList)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	templates, err := template.R().GetAll()
	if err != nil {
		zap.L().Error("Cannot retrieve templates", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	sort.SliceStable(templates, func(i, j int) bool {
		return templates[i].ID < templates[j].ID
	})

	httputil.JSON(w, r, templates)
}

// GetTemplate godoc
//
//	@Id				GetTemplate
//
//	@Summary		Get an email template
//	@Description	Get an email template by ID
//	@Tags			Templates
//	@Produce		json
//	@Param			id	path	string	true	"Template ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	template.Template	"template"
//	@Failure		400	"Status Bad Request"
//	@Failure		404	"Status Not Found"
//	@Router			/engine/templates/{id} [get]
func GetTemplate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idTemplate, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing template id", zap.String("templateID", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeTemplate, strconv.FormatInt(idTemplate, 10), permissions.ActionGet)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	tmpl, err := template.R().Get(idTemplate)
	if err != nil {
		zap.L().Error("Cannot retrieve template", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	httputil.JSON(w, r, tmpl)
}

// GetTemplateByName godoc
//
//	@Id				GetTemplateByName
//
//	@Summary		Get an email template by name
//	@Description	Get an email template by name
//	@Tags			Templates
//	@Produce		json
//	@Param			name	path	string	true	"Template Name"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	template.Template	"template"
//	@Failure		400	"Status Bad Request"
//	@Failure		404	"Status Not Found"
//	@Router			/engine/templates/name/{name} [get]
func GetTemplateByName(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeTemplate, permissions.All, permissions.ActionGet)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	tmpl, err := template.R().GetByName(name)
	if err != nil {
		zap.L().Error("Cannot retrieve template", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	httputil.JSON(w, r, tmpl)
}

// PostTemplate godoc
//
//	@Summary		Create a new email template
//	@Description	Create a new email template
//	@Tags			Templates
//	@Accept			json
//	@Produce		json
//	@Param			template	body	template.Template	true	"Email Template"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	template.Template	"created template with ID"
//	@Failure		400	"Status Bad Request"
//	@Failure		500	"Status Internal Server Error"
//	@Router			/engine/templates [post]
func PostTemplate(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeTemplate, permissions.All, permissions.ActionCreate)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var tmpl template.Template
	err := json.NewDecoder(r.Body).Decode(&tmpl)
	if err != nil {
		zap.L().Warn("Error on unmarshalling template", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}

	if err := tmpl.Validate(); err != nil {
		zap.L().Warn("Template validation failed", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIResourceInvalid, err)
		return
	}

	id, err := template.R().Create(tmpl)
	if err != nil {
		zap.L().Error("Cannot create template", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBInsertFailed, err)
		return
	}

	tmpl.ID = id
	httputil.JSON(w, r, tmpl)
}

// PutTemplate godoc
//
//	@Summary		Update an email template
//	@Description	Update an email template
//	@Tags			Templates
//	@Accept			JSON
//	@Produce		JSON
//	@Param			id			path	string				true	"Template ID"
//	@Param			template	body	template.Template	true	"Email Template"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	template.Template	"updated template"
//	@Failure		400	"Status Bad Request"
//	@Failure		500	"Status Internal Server Error"
//	@Router			/engine/templates/{id} [put]
func PutTemplate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idTemplate, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing template id", zap.String("templateID", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeTemplate, strconv.FormatInt(idTemplate, 10), permissions.ActionUpdate)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var tmpl template.Template
	err = json.NewDecoder(r.Body).Decode(&tmpl)
	if err != nil {
		zap.L().Warn("Error on unmarshalling template", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}

	tmpl.ID = idTemplate
	if err := tmpl.Validate(); err != nil {
		zap.L().Warn("Template validation failed", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIResourceInvalid, err)
		return
	}

	err = template.R().Update(tmpl)
	if err != nil {
		zap.L().Error("Cannot update template", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBUpdateFailed, err)
		return
	}

	httputil.JSON(w, r, tmpl)
}

// DeleteTemplate godoc
//
//	@Summary		Delete an email template
//	@Description	Delete an email template
//	@Tags			Templates
//	@Param			id	path	string	true	"Template ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	"Status OK"
//	@Failure		400	"Status Bad Request"
//	@Failure		500	"Status Internal Server Error"
//	@Router			/engine/templates/{id} [delete]
func DeleteTemplate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idTemplate, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing template id", zap.String("templateID", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeTemplate, strconv.FormatInt(idTemplate, 10), permissions.ActionDelete)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	err = template.R().Delete(idTemplate)
	if err != nil {
		zap.L().Error("Cannot delete template", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBDeleteFailed, err)
		return
	}

	httputil.JSON(w, r, nil)
}
