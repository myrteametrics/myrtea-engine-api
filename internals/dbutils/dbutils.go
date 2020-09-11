package dbutils

import "time"

// DBQueryOptionnal regroups all parameters used to alter an SQL Query with externals parameters
type DBQueryOptionnal struct {
	Limit  int
	Offset int
	MaxAge time.Duration
	// Sorts []string //TODO: real sorting struct (order by)
	// Filters []string //TODO: real filters struct (where)
}
