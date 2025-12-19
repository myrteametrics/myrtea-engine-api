package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/myrteametrics/myrtea-engine-api/v5/internal/tag"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/security/permissions"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/utils/httputil"
	"go.uber.org/zap"
)

// GetTags godoc
//
//	@Id				GetTags
//
//	@Summary		Get all tag definitions
//	@Description	Get all tag definitions
//	@Tags			Tags
//	@Produce		json
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{array}		tag.Tag				"list of all tags"
//	@Failure		500	{object}	httputil.APIError	"Internal Server Error"
//	@Router			/engine/tags [get]
func GetTags(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeTag, permissions.All, permissions.ActionList)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	tags, err := tag.R().GetAll()
	if err != nil {
		zap.L().Error("Error getting tags", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	sort.SliceStable(tags, func(i, j int) bool {
		return tags[i].Id < tags[j].Id
	})

	httputil.JSON(w, r, tags)
}

// GetTag godoc
//
//	@Id				GetTag
//
//	@Summary		Get a tag definition
//	@Description	Get a tag definition
//	@Tags			Tags
//	@Produce		json
//	@Param			id	path	int	true	"Tag ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	tag.Tag				"tag"
//	@Failure		400	{object}	httputil.APIError	"Bad Request"
//	@Router			/engine/tags/{id} [get]
func GetTag(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idTag, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing tag id", zap.String("tagID", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeTag, strconv.FormatInt(idTag, 10), permissions.ActionGet)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	t, found, err := tag.R().Get(idTag)
	if err != nil {
		zap.L().Error("Cannot get tag", zap.Int64("tagId", idTag), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("Tag does not exist", zap.Int64("tagId", idTag))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, err)
		return
	}

	httputil.JSON(w, r, t)
}

// ValidateTag godoc
//
//	@Id				ValidateTag
//
//	@Summary		Validate a new tag definition
//	@Description	Validate a new tag definition
//	@Tags			Tags
//	@Accept			json
//	@Produce		json
//	@Param			tag	body	tag.Tag	true	"Tag definition (json)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	tag.Tag				"tag"
//	@Failure		400	{object}	httputil.APIError	"Bad Request"
//	@Failure		500	{object}	httputil.APIError	"Internal Server Error"
//	@Router			/engine/tags/validate [post]
func ValidateTag(w http.ResponseWriter, r *http.Request) {
	var newTag tag.Tag
	err := json.NewDecoder(r.Body).Decode(&newTag)
	if err != nil {
		zap.L().Warn("Tag json decoding", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}

	if ok, err := newTag.IsValid(); !ok {
		zap.L().Warn("Tag is not valid", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIResourceInvalid, err)
		return
	}

	httputil.JSON(w, r, newTag)
}

// PostTag godoc
//
//	@Id				PostTag
//
//	@Summary		Create a new tag definition
//	@Description	Create a new tag definition
//	@Tags			Tags
//	@Accept			json
//	@Produce		json
//	@Param			tag	body	tag.Tag	true	"Tag definition (json)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	tag.Tag				"tag"
//	@Failure		400	{object}	httputil.APIError	"Bad Request"
//	@Failure		500	{object}	httputil.APIError	"Internal Server Error"
//	@Router			/engine/tags [post]
func PostTag(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeTag, permissions.All, permissions.ActionCreate)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var newTag tag.Tag
	err := json.NewDecoder(r.Body).Decode(&newTag)
	if err != nil {
		zap.L().Warn("Tag json decoding", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}

	if ok, err := newTag.IsValid(); !ok {
		zap.L().Warn("Tag is not valid", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIResourceInvalid, err)
		return
	}

	newTagID, err := tag.R().Create(newTag)
	if err != nil {
		zap.L().Error("Error while creating the Tag", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBInsertFailed, err)
		return
	}

	newTagGet, found, err := tag.R().Get(newTagID)
	if err != nil {
		zap.L().Error("Cannot get tag", zap.Int64("tagId", newTagID), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("Tag does not exist after creation", zap.Int64("tagId", newTagID))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFoundAfterInsert, err)
		return
	}

	httputil.JSON(w, r, newTagGet)
}

// PutTag godoc
//
//	@Id				PutTag
//
//	@Summary		Create or replace a tag definition
//	@Description	Create or replace a tag definition
//	@Tags			Tags
//	@Accept			json
//	@Produce		json
//	@Param			id	path	int		true	"Tag ID"
//	@Param			tag	body	tag.Tag	true	"Tag definition (json)"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	tag.Tag				"tag"
//	@Failure		400	{object}	httputil.APIError	"Bad Request"
//	@Failure		500	{object}	httputil.APIError	"Internal Server Error"
//	@Router			/engine/tags/{id} [put]
func PutTag(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idTag, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing tag id", zap.String("tagID", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeTag, strconv.FormatInt(idTag, 10), permissions.ActionUpdate)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var updatedTag tag.Tag
	err = json.NewDecoder(r.Body).Decode(&updatedTag)
	if err != nil {
		zap.L().Warn("Tag json decoding", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		return
	}
	updatedTag.Id = idTag

	if ok, err := updatedTag.IsValid(); !ok {
		zap.L().Warn("Tag is not valid", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIResourceInvalid, err)
		return
	}

	err = tag.R().Update(updatedTag)
	if err != nil {
		zap.L().Error("Error while updating the Tag", zap.Int64("idTag", idTag), zap.Any("tag", updatedTag), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBUpdateFailed, err)
		return
	}

	updatedTagGet, found, err := tag.R().Get(idTag)
	if err != nil {
		zap.L().Error("Cannot get tag", zap.Int64("tagId", idTag), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("Tag does not exist after update", zap.Int64("tagId", idTag))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, err)
		return
	}

	httputil.JSON(w, r, updatedTagGet)
}

// DeleteTag godoc
//
//	@Id				DeleteTag
//
//	@Summary		Delete a tag definition
//	@Description	Delete a tag definition
//	@Tags			Tags
//	@Produce		json
//	@Param			id	path	int	true	"Tag ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	"Status OK"
//	@Failure		400	{object}	httputil.APIError	"Bad Request"
//	@Router			/engine/tags/{id} [delete]
func DeleteTag(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	idTag, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing tag id", zap.String("tagID", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeTag, strconv.FormatInt(idTag, 10), permissions.ActionDelete)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	err = tag.R().Delete(idTag)
	if err != nil {
		zap.L().Error("Error while deleting the Tag", zap.String("Tag ID", id), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBDeleteFailed, err)
		return
	}

	httputil.OK(w, r)
}

// GetTagsBySituation godoc
//
//	@Id				GetTagsBySituation
//
//	@Summary		Get all tags for a situation
//	@Description	Get all tags associated with a specific situation
//	@Tags			Tags
//	@Produce		json
//	@Param			situationId	path	int	true	"Situation ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{array}		tag.Tag				"list of tags"
//	@Failure		400	{object}	httputil.APIError	"Bad Request"
//	@Failure		500	{object}	httputil.APIError	"Internal Server Error"
//	@Router			/engine/tags/situations/{situationId} [get]
func GetTagsBySituation(w http.ResponseWriter, r *http.Request) {
	situationId := chi.URLParam(r, "situationId")
	idSituation, err := strconv.ParseInt(situationId, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing situation id", zap.String("situationID", situationId), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeTagSituation, permissions.All, permissions.ActionList)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	tags, err := tag.R().GetTagsBySituationId(idSituation)
	if err != nil {
		zap.L().Error("Error getting tags for situation", zap.Int64("situationId", idSituation), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	httputil.JSON(w, r, tags)
}

// AddTagToSituation godoc
//
//	@Id				AddTagToSituation
//
//	@Summary		Associate a tag with a situation
//	@Description	Create a link between a tag and a situation
//	@Tags			Tags
//	@Produce		json
//	@Param			situationId	path	int	true	"Situation ID"
//	@Param			tagId		path	int	true	"Tag ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	"Status OK"
//	@Failure		400	{object}	httputil.APIError	"Bad Request"
//	@Failure		500	{object}	httputil.APIError	"Internal Server Error"
//	@Router			/engine/tags/{tagId}/situations/{situationId} [post]
func AddTagToSituation(w http.ResponseWriter, r *http.Request) {
	situationId := chi.URLParam(r, "situationId")
	idSituation, err := strconv.ParseInt(situationId, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing situation id", zap.String("situationID", situationId), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	tagId := chi.URLParam(r, "tagId")
	idTag, err := strconv.ParseInt(tagId, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing tag id", zap.String("tagID", tagId), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeTagSituation, strconv.FormatInt(idSituation, 10), permissions.ActionUpdate)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	// Check if tag exists
	_, found, err := tag.R().Get(idTag)
	if err != nil {
		zap.L().Error("Cannot get tag", zap.Int64("tagId", idTag), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("Tag does not exist", zap.Int64("tagId", idTag))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, errors.New("tag not found"))
		return
	}

	err = tag.R().CreateLinkWithSituation(idTag, idSituation)
	if err != nil {
		zap.L().Error("Error creating link between tag and situation",
			zap.Int64("tagId", idTag),
			zap.Int64("situationId", idSituation),
			zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBInsertFailed, err)
		return
	}

	httputil.OK(w, r)
}

// RemoveTagFromSituation godoc
//
//	@Id				RemoveTagFromSituation
//
//	@Summary		Remove tag association from a situation
//	@Description	Delete the link between a tag and a situation
//	@Tags			Tags
//	@Produce		json
//	@Param			situationId	path	int	true	"Situation ID"
//	@Param			tagId		path	int	true	"Tag ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	"Status OK"
//	@Failure		400	{object}	httputil.APIError	"Bad Request"
//	@Failure		500	{object}	httputil.APIError	"Internal Server Error"
//	@Router			/engine/tags/{tagId}/situations/{situationId} [delete]
func RemoveTagFromSituation(w http.ResponseWriter, r *http.Request) {
	situationId := chi.URLParam(r, "situationId")
	idSituation, err := strconv.ParseInt(situationId, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing situation id", zap.String("situationID", situationId), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	tagId := chi.URLParam(r, "tagId")
	idTag, err := strconv.ParseInt(tagId, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing tag id", zap.String("tagID", tagId), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeTagSituation, strconv.FormatInt(idSituation, 10), permissions.ActionUpdate)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	err = tag.R().DeleteLinkWithSituation(idTag, idSituation)
	if err != nil {
		zap.L().Error("Error deleting link between tag and situation",
			zap.Int64("tagId", idTag),
			zap.Int64("situationId", idSituation),
			zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBDeleteFailed, err)
		return
	}

	httputil.OK(w, r)
}

// GetTagsByTemplateInstance godoc
//
//	@Id				GetTagsByTemplateInstance
//
//	@Summary		Get all tags for a template instance
//	@Description	Get all tags associated with a specific template instance
//	@Tags			Tags
//	@Produce		json
//	@Param			instanceId	path	int	true	"Template Instance ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{array}		tag.Tag				"list of tags"
//	@Failure		400	{object}	httputil.APIError	"Bad Request"
//	@Failure		500	{object}	httputil.APIError	"Internal Server Error"
//	@Router			/engine/tags/situationinstances/{instanceId} [get]
func GetTagsByTemplateInstance(w http.ResponseWriter, r *http.Request) {
	instanceId := chi.URLParam(r, "instanceId")
	idTemplateInstance, err := strconv.ParseInt(instanceId, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing template instance id", zap.String("templateInstanceID", instanceId), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	// Assuming you have a permission type for template instances
	if !userCtx.HasPermission(permissions.New(permissions.TypeTagSituationInstance, permissions.All, permissions.ActionList)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	tags, err := tag.R().GetTagsByTemplateInstanceId(idTemplateInstance)
	if err != nil {
		zap.L().Error("Error getting tags for template instance", zap.Int64("instanceId", idTemplateInstance), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	httputil.JSON(w, r, tags)
}

// AddTagToTemplateInstance godoc
//
//	@Id				AddTagToTemplateInstance
//
//	@Summary		Associate a tag with a template instance
//	@Description	Create a link between a tag and a template instance
//	@Tags			Tags
//	@Produce		json
//	@Param			instanceId	path	int	true	"Template Instance ID"
//	@Param			tagId		path	int	true	"Tag ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	"Status OK"
//	@Failure		400	{object}	httputil.APIError	"Bad Request"
//	@Failure		500	{object}	httputil.APIError	"Internal Server Error"
//	@Router			/engine/tags/{tagId}/situationinstances/{instanceId} [post]
func AddTagToTemplateInstance(w http.ResponseWriter, r *http.Request) {
	instanceId := chi.URLParam(r, "instanceId")
	idTemplateInstance, err := strconv.ParseInt(instanceId, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing template instance id", zap.String("templateInstanceID", instanceId), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	tagId := chi.URLParam(r, "tagId")
	idTag, err := strconv.ParseInt(tagId, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing tag id", zap.String("tagID", tagId), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeTagSituationInstance, strconv.FormatInt(idTemplateInstance, 10), permissions.ActionUpdate)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	// Check if tag exists
	_, found, err := tag.R().Get(idTag)
	if err != nil {
		zap.L().Error("Cannot get tag", zap.Int64("tagId", idTag), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}
	if !found {
		zap.L().Warn("Tag does not exist", zap.Int64("tagId", idTag))
		httputil.Error(w, r, httputil.ErrAPIDBResourceNotFound, errors.New("tag not found"))
		return
	}

	err = tag.R().CreateLinkWithTemplateInstance(idTag, idTemplateInstance)
	if err != nil {
		zap.L().Error("Error creating link between tag and template instance",
			zap.Int64("tagId", idTag),
			zap.Int64("instanceId", idTemplateInstance),
			zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBInsertFailed, err)
		return
	}

	httputil.OK(w, r)
}

// RemoveTagFromTemplateInstance godoc
//
//	@Id				RemoveTagFromTemplateInstance
//
//	@Summary		Remove tag association from a template instance
//	@Description	Delete the link between a tag and a template instance
//	@Tags			Tags
//	@Produce		json
//	@Param			instanceId	path	int	true	"Template Instance ID"
//	@Param			tagId		path	int	true	"Tag ID"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	"Status OK"
//	@Failure		400	{object}	httputil.APIError	"Bad Request"
//	@Failure		500	{object}	httputil.APIError	"Internal Server Error"
//	@Router			/engine/tags/{tagId}/situationinstances/{instanceId} [delete]
func RemoveTagFromTemplateInstance(w http.ResponseWriter, r *http.Request) {
	instanceId := chi.URLParam(r, "instanceId")
	idTemplateInstance, err := strconv.ParseInt(instanceId, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing template instance id", zap.String("templateInstanceID", instanceId), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	tagId := chi.URLParam(r, "tagId")
	idTag, err := strconv.ParseInt(tagId, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing tag id", zap.String("tagID", tagId), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeTagSituationInstance, strconv.FormatInt(idTemplateInstance, 10), permissions.ActionUpdate)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	err = tag.R().DeleteLinkWithTemplateInstance(idTag, idTemplateInstance)
	if err != nil {
		zap.L().Error("Error deleting link between tag and template instance",
			zap.Int64("tagId", idTag),
			zap.Int64("instanceId", idTemplateInstance),
			zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBDeleteFailed, err)
		return
	}

	httputil.OK(w, r)
}

// GetAllSituationsTags godoc
//
//	@Id				GetAllSituationsTags
//
//	@Summary		Get all tags for all situations
//	@Description	Get all tags grouped by situation ID
//	@Tags			Tags
//	@Produce		json
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	map[string][]tag.Tag	"map of situation IDs to tags"
//	@Failure		500	{object}	httputil.APIError		"Internal Server Error"
//	@Router			/engine/tags/situations [get]
func GetAllSituationsTags(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeTag, permissions.All, permissions.ActionList)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	situationsTags, err := tag.R().GetSituationsTags()
	if err != nil {
		zap.L().Error("Error getting all situations tags", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	// Convert map[int64][]Tag to map[string][]Tag for JSON rendering
	result := make(map[string][]tag.Tag)
	for situationID, tags := range situationsTags {
		result[strconv.FormatInt(situationID, 10)] = tags
	}

	httputil.JSON(w, r, result)
}

// GetSituationTemplateInstanceTags godoc
//
//	@Id				GetSituationTemplateInstanceTags
//
//	@Summary		Get all tags for template instances in a situation
//	@Description	Get all tags associated with template instances in a specific situation
//	@Tags			Tags
//	@Produce		json
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Param			situationId	path		string					true	"Situation ID"
//	@Success		200			{object}	map[string][]tag.Tag	"map of template instance IDs to tags"
//	@Failure		500			"internal server error"
//	@Router			/engine/tags/situations/{situationId}/instances [get]
func GetSituationTemplateInstanceTags(w http.ResponseWriter, r *http.Request) {
	situationId := chi.URLParam(r, "situationId")
	idSituation, err := strconv.ParseInt(situationId, 10, 64)
	if err != nil {
		zap.L().Warn("Error on parsing situation id", zap.String("situationID", situationId), zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIParsingInteger, err)
		return
	}

	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeTag, permissions.All, permissions.ActionList)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	templateInstancesTags, err := tag.R().GetSituationInstanceTags(idSituation)
	if err != nil {
		zap.L().Error("Error getting all template instances tags", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDBSelectFailed, err)
		return
	}

	// Convert map[int64][]Tag to map[string][]Tag for JSON rendering
	result := make(map[string][]tag.Tag)
	for situationID, tags := range templateInstancesTags {
		result[strconv.FormatInt(situationID, 10)] = tags
	}

	httputil.JSON(w, r, templateInstancesTags)
}
