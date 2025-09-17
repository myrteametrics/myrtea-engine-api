package handler

import (
	"fmt"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/security/permissions"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/utils/httputil"
	"github.com/myrteametrics/myrtea-sdk/v5/expression"
	"go.uber.org/zap"
	"net/http"

	"encoding/json"
	"errors"
)

// EvaluateExpression godoc
//
//	@Summary		EvaluateExpression an expression with variables
//	@Description	EvaluateExpression an expression with variables and return the result
//	@Tags			Expressions
//	@Accept			json
//	@Produce		json
//	@Param			request	body	any	true	"Expression and variables"
//	@Security		Bearer
//	@Security		ApiKeyAuth
//	@Success		200	{object}	string	"evaluated expression"
//	@Failure		400	"Status Bad Request"
//	@Failure		500	"Status Internal Server Error"
//	@Router			/engine/expression/evaluate [post]
func EvaluateExpression(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := GetUserFromContext(r)
	if !userCtx.HasPermission(permissions.New(permissions.TypeTemplate, permissions.All, permissions.ActionGet)) {
		httputil.Error(w, r, httputil.ErrAPISecurityNoPermissions, errors.New("missing permission"))
		return
	}

	var request map[string]interface{}
	var body interface{} = ""

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		zap.L().Error("Error on unmarshalling evaluation request", zap.Error(err))
		httputil.Error(w, r, httputil.ErrAPIDecodeJSONBody, err)
		body = "Error on unmarshalling evaluation request"
	}

	exprInterface, ok := request["expression"]
	if !ok {
		body = "Invalid parameter expression"
	}
	expr, ok := exprInterface.(string)
	if !ok {
		body = "Invalid parameter expression"
	}

	varsInterface, ok := request["variables"]
	if !ok {
		body = "Invalid parameter variables"
	}
	vars, ok := varsInterface.(map[string]interface{})
	if !ok {
		body = "Invalid parameter variables"
	}

	if body == "" {
		body, err = expression.Process(expression.LangEval, expr, vars)
		if err != nil {
			zap.L().Error("Error evaluating expression", zap.Error(err), zap.String("expression", expr), zap.Any("variables", vars))
			body = "Unable to evaluate this expression"
		}
	}

	response := map[string]interface{}{}
	response["body"] = fmt.Sprintf("%v", body)
	httputil.JSON(w, r, response)
	return
}
