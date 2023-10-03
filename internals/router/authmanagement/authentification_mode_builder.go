package authmanagement

import (
	sq "github.com/Masterminds/squirrel"
)

type AuthenticationModeBuilder struct{}

func (builder AuthenticationModeBuilder) newStatement() sq.StatementBuilderType {
	return sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
}

func (builder AuthenticationModeBuilder) SelectMode() sq.SelectBuilder {
	return builder.newStatement().
		Select("mode").
		From("AuthenticationMode").
		Where(sq.Eq{"id": 1}).
		Limit(1)
}
