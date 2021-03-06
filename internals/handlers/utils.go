package handlers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"net/http"

	"github.com/myrteametrics/myrtea-engine-api/v4/internals/groups"
	"github.com/myrteametrics/myrtea-engine-api/v4/internals/models"
	"go.uber.org/zap"
)

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

// ParsePlaceholders parse multiple key:value entries separated by commas from a string
func ParsePlaceholders(pStr string) (map[string]string, error) {
	placeholders := make(map[string]string)
	if pStr != "" {
		rawPlaceholders := strings.Split(pStr, ",")
		for _, rawPlaceholder := range rawPlaceholders {
			keyValue := strings.Split(rawPlaceholder, ":")
			if len(keyValue) != 2 {
				return nil, fmt.Errorf("Invalid placeholder key-value : %s", keyValue)
			}
			placeholders[keyValue[0]] = keyValue[1]
		}
	}
	return placeholders, nil
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
func GetUserFromContext(r *http.Request) *groups.UserWithGroups {
	c := r.Context()
	_user := c.Value(models.ContextKeyUser)
	if _user == nil {
		zap.L().Warn("No context user provided")
		return nil
	}
	user := _user.(groups.UserWithGroups)
	return &user
}

// GetUserGroupsFromUser extracts group IDs from a user
func GetUserGroupsFromUser(user *groups.UserWithGroups) []int64 {
	groupsID := make([]int64, 0)
	for _, group := range user.Groups {
		groupsID = append(groupsID, group.ID)
	}
	return groupsID
}

// GetUserGroupsFromContext extract the logged user groups from the request context
func GetUserGroupsFromContext(r *http.Request) []int64 {
	user := GetUserFromContext(r)
	if user == nil {
		return nil
	}
	return GetUserGroupsFromUser(user)
}
