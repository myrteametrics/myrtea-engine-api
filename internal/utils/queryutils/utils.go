package queryutils

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/myrteametrics/myrtea-engine-api/v5/internal/models"
)

const (
	defaultLimit  = 50
	defaultOffset = 0
)

var (
	checkSQLFieldRegex = regexp.MustCompile("^[A-Za-z0-9_]+$")
)

// SecurityCheckSQLField Check if a string is only composed of alphanumeric or underscore
func securityCheckSQLField(str string) bool {
	return checkSQLFieldRegex.MatchString(str)
}

// AppendSearchOptions alter the input query and parameters based on the input search options
func AppendSearchOptions(query string, params map[string]interface{}, options models.SearchOptions, sortFieldPrefix string) (string, map[string]interface{}, error) {
	if nbSortBy := len(options.SortBy); nbSortBy > 0 {
		query += ` ORDER BY `
		for i, sortBy := range options.SortBy {
			// The ORDER BY clause cannot be used with SQL parameter bindings
			// To ensure security and prevent SQLi, we make a common regex check on the field content
			if !securityCheckSQLField(sortBy.Field) {
				return "", nil, fmt.Errorf("unsafe sort field detected : '%s'", sortBy.Field)
			}
			query += fmt.Sprintf("%s.%s %s", sortFieldPrefix, sortBy.Field, strings.ToUpper(sortBy.Order.String()))
			if i < nbSortBy-1 {
				query += `, `
			}
		}
	}

	query += ` LIMIT :limit`
	if options.Limit > 0 {
		params["limit"] = options.Limit
	} else {
		params["limit"] = defaultLimit
	}

	query += ` OFFSET :offset`
	if options.Offset > 0 {
		params["offset"] = options.Offset
	} else {
		params["offset"] = defaultOffset
	}

	return query, params, nil
}
