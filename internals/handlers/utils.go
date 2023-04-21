package handlers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"net/http"

	"github.com/myrteametrics/myrtea-engine-api/v5/internals/models"
	"github.com/myrteametrics/myrtea-engine-api/v5/internals/security/users"
	"go.uber.org/zap"
)

func QueryParamToOptionalInt(r *http.Request, name string, orDefault int) (int, error) {
	param := r.URL.Query().Get(name)
	if param != "" {
		return strconv.Atoi(param)
	}
	return orDefault, nil
}

func QueryParamToOptionalInt64(r *http.Request, name string, orDefault int64) (int64, error) {
	param := r.URL.Query().Get(name)
	if param != "" {
		return strconv.ParseInt(param, 10, 64)
	}
	return orDefault, nil
}

func QueryParamToOptionalStringArray(r *http.Request, name string, separator string, orDefault []string) []string {
	param := r.URL.Query().Get(name)
	if param != "" {
		return strings.Split(param, separator)
	}
	return orDefault
}

func QueryParamToOptionalTime(r *http.Request, name string, orDefault time.Time) (time.Time, error) {
	param := r.URL.Query().Get(name)
	if param != "" {
		return time.Parse("2006-01-02T15:04:05.000Z07:00", param)
	}
	return orDefault, nil
}

func QueryParamToTime(r *http.Request, name string) (time.Time, error) {
	param := r.URL.Query().Get(name)
	if param != "" {
		return time.Parse("2006-01-02T15:04:05.000Z07:00", param)
	}
	return time.Time{}, fmt.Errorf("missing query parameter %s", name)
}

func QueryParamToOptionalDuration(r *http.Request, name string, orDefault time.Duration) (time.Duration, error) {
	param := r.URL.Query().Get(name)
	if param != "" {
		return time.ParseDuration(param)
	}
	return orDefault, nil
}

func QueryParamToOptionalBool(r *http.Request, name string, orDefault bool) (bool, error) {
	param := r.URL.Query().Get(name)
	if param != "" {
		return strconv.ParseBool(param)
	}
	return orDefault, nil
}

// QueryParamToOptionalKeyValues parse multiple key:value entries separated by commas from a string
func QueryParamToOptionalKeyValues(r *http.Request, name string, orDefault map[string]string) (map[string]string, error) {
	param := r.URL.Query().Get(name)
	if param != "" {
		keyValues := make(map[string]string)
		rawKeyValues := strings.Split(param, ",")
		for _, rawKeyValue := range rawKeyValues {
			keyValue := strings.Split(rawKeyValue, ":")
			if len(keyValue) != 2 {
				return nil, fmt.Errorf("invalid placeholder key-value : %s", keyValue)
			}
			keyValues[keyValue[0]] = keyValue[1]
		}
		return keyValues, nil
	}
	return orDefault, nil
}

// ParseTime try to parse a supposed time string as a time.Time or returns time.Now()
func ParseTime(tStr string) (time.Time, error) {
	t, err := time.Parse("2006-01-02T15:04:05.000Z07:00", tStr)
	if err != nil {
		return time.Now().UTC(), err
	}
	return t, nil
}

// ParseInt try to parse a string as an int or returns 0
func ParseInt(tStr string) (int, error) {
	if tStr != "" {
		i, err := strconv.Atoi(tStr)
		if err != nil {
			return 0, err
		}
		return i, nil
	}
	return 0, nil
}

// ParseDuration try to parse a string as an int or returns 0
func ParseDuration(dStr string) (time.Duration, error) {
	if dStr != "" {
		d, err := time.ParseDuration(dStr)
		if err != nil {
			return 0, err
		}
		return d, nil
	}
	return 0, nil
}

// sortByRegex is a regex matching expression similar to <order>(<field>)
// where <order> must be 'asc' or 'desc' and <field> cannot contains parenthesis
var sortByRegex = regexp.MustCompile(`^(asc|desc)\(([A-Za-z0-9_]+?)\)$`)

// ParseSortBy parse multiple <order>(<field>) entries separated by commas from a string
func ParseSortBy(rawSortByStr string, allowedFields []string) ([]models.SortOption, error) {
	sortOptions := make([]models.SortOption, 0)
	for _, sortByStr := range strings.Split(rawSortByStr, ",") {
		sortByStr = strings.TrimSpace(sortByStr)
		if sortByStr == "" {
			continue
		}

		parsing := sortByRegex.FindStringSubmatch(sortByStr)
		if len(parsing) < 3 {
			return nil, fmt.Errorf("invalid sortby clause '%s'", sortByStr)
		}

		order := models.ToSortOptionOrder(parsing[1])
		if order == 0 {
			return nil, fmt.Errorf("invalid sortby order found '%s'", parsing[1])
		}

		field := parsing[2]
		found := false
		for _, allowedField := range allowedFields {
			if field == allowedField {
				found = true
				break
			}
		}

		if !found {
			return nil, fmt.Errorf("invalid sortby field found '%s'", field)
		}

		sortOptions = append(sortOptions, models.SortOption{
			Field: field,
			Order: order,
		})
	}
	return sortOptions, nil
}

// GetUserFromContext extract the logged user from the request context
func GetUserFromContext(r *http.Request) (users.UserWithPermissions, bool) {
	c := r.Context()
	_user := c.Value(models.ContextKeyUser)
	if _user == nil {
		zap.L().Warn("No context user provided")
		return users.UserWithPermissions{}, false
	}
	user := _user.(users.UserWithPermissions)
	return user, true
}
