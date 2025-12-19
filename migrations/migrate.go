package migrations

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
)

//go:embed *.sql
var embedMigrations embed.FS

func Migrate(db *sql.DB) error {
	goose.SetBaseFS(embedMigrations)
	goose.SetLogger(&customLogger{})

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	return goose.Up(db, ".")
}

type customLogger struct{}

func (l *customLogger) Printf(format string, v ...interface{}) {
	// Custom log format
	zap.L().Info(fmt.Sprintf(format, v...))
}

func (l *customLogger) Fatalf(format string, v ...interface{}) {
	zap.L().Error(fmt.Sprintf(format, v...))
}
