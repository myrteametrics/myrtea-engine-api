package dbutils

import (
	"errors"
	"github.com/lib/pq"
	"time"
)

// DBQueryOptionnal regroups all parameters used to alter an SQL Query with externals parameters
type DBQueryOptionnal struct {
	Limit  int
	Offset int
	MaxAge time.Duration
	// Sorts []string // TODO : real sorting struct (order by)
	// Filters []string // TODO : real filters struct (where)
}

// UniqueViolation checks if the error is of code 23505
func UniqueViolation(err error) *pq.Error {
	var pqerr *pq.Error
	if errors.As(err, &pqerr) &&
		pqerr.Code == "23505" {
		return pqerr
	}
	return nil
}
