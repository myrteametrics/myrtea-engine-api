package connector

import (
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// PostgresRepository is a repository containing the Fact definition based on a PSQL database and
// implementing the repository interface
type PostgresRepository struct {
	conn *sqlx.DB
}

// NewPostgresRepository returns a new instance of PostgresRepository
func NewPostgresRepository(dbClient *sqlx.DB) Repository {
	r := PostgresRepository{
		conn: dbClient,
	}
	var ifm Repository = &r
	return ifm
}

//GetLastConnectionReading returns the datetime of the last connections reading for a connector id
func (r *PostgresRepository) GetLastConnectionReading(connectorID string, successOnly bool) (map[string]time.Time, error) {
	query := `SELECT DISTINCT ON (name) name, ts 
				FROM connectors_executions_log_v1 
				WHERE connector_id = :connector_id AND (:success_only = false OR success = :success_only)
				ORDER BY name, ts DESC`

	rows, err := r.conn.NamedQuery(query, map[string]interface{}{
		"connector_id": connectorID,
		"success_only": successOnly,
	})
	if err != nil {
		zap.L().Error("Couldn't retrieve the Connections reading datetime", zap.Error(err))
		return nil, errors.New("couldn't retrieve the connections reading " + err.Error())
	}
	defer rows.Close()

	lastReading := make(map[string]time.Time, 0)
	for rows.Next() {
		var name string
		var ts time.Time

		err = rows.Scan(&name, &ts)
		if err != nil {
			zap.L().Error("Couldn't read the rows:", zap.Error(err))
			return nil, errors.New("couldn't read the rows: " + err.Error())
		}

		lastReading[name] = ts
	}

	return lastReading, nil
}
