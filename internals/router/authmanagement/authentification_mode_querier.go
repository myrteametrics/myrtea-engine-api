package authmanagement

import (
	"database/sql"
	"errors"
	"sync"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

type AuthenticationMode struct {
	Mode string `json:"mode"`
}

type AuthenticationModeQuerier struct {
	Builder AuthenticationModeBuilder
	Conn    *sqlx.DB
}

var (
	_globalQuerierMu    sync.RWMutex
	_globalQuerier      *AuthenticationModeQuerier
)



func New(db *sqlx.DB) *AuthenticationModeQuerier {
	querier := &AuthenticationModeQuerier{
		Conn:    db,
		Builder: AuthenticationModeBuilder{},
	}
	return querier
}


func S() *AuthenticationModeQuerier {
	_globalQuerierMu.RLock()
	defer _globalQuerierMu.RUnlock()

	return _globalQuerier
}

func ReplaceGlobals(querier *AuthenticationModeQuerier) func() {
	_globalQuerierMu.Lock()
	defer _globalQuerierMu.Unlock()

	prev := _globalQuerier
	_globalQuerier = querier

	return func() {
		ReplaceGlobals(prev)
	}
}

func (querier AuthenticationModeQuerier) SetMode(mode string) error {
	_, err := querier.Conn.Exec(`INSERT INTO AuthenticationMode (id, mode) VALUES (1, $1) ON CONFLICT (id) DO UPDATE SET mode = EXCLUDED.mode;`, mode)
	return err
}

func (querier AuthenticationModeQuerier) Query(builder sq.SelectBuilder) (AuthenticationMode, error) {
	rows, err := builder.RunWith(querier.Conn.DB).Query()
	if err != nil {
		return AuthenticationMode{}, err
	}
	defer rows.Close()

    if rows.Next() {
        return querier.scan(rows)
    }

    return AuthenticationMode{}, errors.New("no data found")
}

func (querier AuthenticationModeQuerier) scan(rows *sql.Rows) (AuthenticationMode, error) {
	mode := AuthenticationMode{}

	err := rows.Scan(&mode.Mode)
	if err != nil {
		return AuthenticationMode{}, errors.New("couldn't scan the retrieved data: " + err.Error())
	}

	return mode, nil
}
