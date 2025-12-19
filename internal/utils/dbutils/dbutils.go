package dbutils

import (
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

var ErrNoRowDeleted = errors.New("no row deleted (or multiple row deleted) instead of 1 row")

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

// CheckRowAffected checks if the number of rows affected by a query is the expected number
func CheckRowAffected(result sql.Result, nbRows int64) error {
	i, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if i != nbRows {
		return ErrNoRowDeleted
	}
	return nil
}

// ScanFirst scans the first row of a sql.Rows and returns the result
func ScanFirst[T any](rows *sql.Rows, scan func(rows *sql.Rows) (T, error)) (T, bool, error) {
	if rows.Next() {
		obj, err := scan(rows)
		return obj, err == nil, err
	}
	var a T
	return a, false, nil
}

// ScanAll scans all the rows of the given rows and returns a slice of DataSource
func ScanAll[T any](rows *sql.Rows, scan func(rows *sql.Rows) (T, error)) ([]T, error) {
	objs := make([]T, 0)
	for rows.Next() {
		obj, err := scan(rows)
		if err != nil {
			zap.L().Warn("scan error", zap.Error(err))
			return []T{}, err
		}
		objs = append(objs, obj)
	}
	return objs, nil
}

// ScanFirstStruct scans the first row of a sql.Rows and returns the result
func ScanFirstStruct[T any](rows *sql.Rows) (T, bool, error) {
	var results []T
	err := sqlx.StructScan(rows, &results)
	if err != nil {
		var a T
		return a, false, err
	}

	if len(results) == 0 {
		var a T
		return a, false, nil
	}

	return results[0], true, nil
}

// ScanAllStruct scans all the rows of the given rows and returns a slice of DataSource
func ScanAllStruct[T any](rows *sql.Rows) ([]T, error) {
	var results []T
	err := sqlx.StructScan(rows, &results)
	if err != nil {
		return []T{}, err
	}
	if results == nil {
		results = make([]T, 0)
	}
	return results, nil
}
