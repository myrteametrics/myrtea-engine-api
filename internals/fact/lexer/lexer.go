package lexer

import (
	"sync"

	parsec "github.com/prataprc/goparsec"
)

// Lexer is a specific lexing component to parse fact condition DSL
type Lexer struct {
	Ast    *parsec.AST
	Parser parsec.Parser
}

var (
	_globalMu    sync.RWMutex
	_globalLexer *Lexer
)

// L is used to access the global lexer singleton
func L() *Lexer {
	_globalMu.RLock()
	defer _globalMu.RUnlock()

	lexer := _globalLexer
	return lexer
}

// ReplaceGlobals affect a new repository to the global lexer singleton
func ReplaceGlobals(lexer *Lexer) func() {
	_globalMu.Lock()
	defer _globalMu.Unlock()

	prev := _globalLexer
	_globalLexer = lexer
	return func() { ReplaceGlobals(prev) }
}

// New creates a new lexer instance
func New(entities []string) (*Lexer, error) {
	ast := parsec.NewAST("AST", 100)

	// Terminal
	tokenEqual := parsec.Token("=", "EQUALS")
	tokenLT := parsec.Token("<", "LESS_THAN")
	tokenGT := parsec.Token(">", "GREATER_THAN")
	tokenGTE := parsec.Token(">=", "GREATER_THAN_OR_EQUALS")
	tokenLTE := parsec.Token("<=", "LESS_THAN_OR_EQUALS")

	tokenAND := parsec.Token("AND", "AND")
	tokenOR := parsec.Token("OR", "OR")

	tokenEXISTS := parsec.Token("EXISTS", "EXISTS")
	tokenOpenScript := parsec.Token("\\${", "OPEN_SCRIPT")
	tokenScriptContent := parsec.Token("[\\w.\\s><=+\\-*\\/\\(\\)]*", "SCRIPT_CONTENT")
	tokenCloseScript := parsec.Token("}", "CLOSE_SCRIPT")

	tokenValue := parsec.Token(`[a-zA-Z0-9\-]+`, "VALUE")
	tokenVariable := parsec.Token(`^\$[a-zA-Z0-9\-]+`, "VARIABLE")

	tokenDate := parsec.Token(`[0-9]{4}-[0-9]{2}-[0-9]{2}(T[0-9]{2}:[0-9]{2}:[0-9]{2}(\.[0-9]{3})?)?`, "DATE")
	tokenNumber := parsec.Token(`[0-9]*[0-9]\.?[0-9]*`, "NUMERIC")

	entityTokens := make([]interface{}, 0)
	for _, entity := range entities {
		entityTokens = append(entityTokens, parsec.Token(entity, "ENTITY"))
	}
	tokenEntity := ast.OrdChoice("entity", nil, entityTokens...)

	boolean := ast.OrdChoice("boolean", nil, tokenAND, tokenOR)
	values := ast.OrdChoice("value", nil, tokenDate, tokenNumber, tokenValue, tokenVariable)
	compareOps := ast.OrdChoice("compare_ops", nil, tokenEqual, tokenLT, tokenGT, tokenGTE, tokenLTE)
	rightBetween := ast.OrdChoice("rightBetween", nil, tokenLT, tokenLTE)
	leftBetween := ast.OrdChoice("leftBetween", nil, tokenGT, tokenGTE)

	rBetween := ast.And("BETWEEN", nil, values, rightBetween, tokenEntity, rightBetween, values)
	lBetween := ast.And("BETWEEN", nil, values, leftBetween, tokenEntity, leftBetween, values)
	compare := ast.And("COMPARE", nil, tokenEntity, compareOps, values) //TODO: Do we need entities here too?
	exists := ast.And("EXISTS", nil, tokenEXISTS, tokenEntity)
	script := ast.And("SCRIPT", nil, tokenOpenScript, tokenScriptContent, tokenCloseScript)

	condition := ast.OrdChoice("CONDITION", nil, exists, rBetween, lBetween, compare, script)
	conditions := ast.Many("CONDITIONS", nil, condition)
	boolCondition := ast.And("BOOLEAN_CONDITION", nil, boolean, conditions)

	expression := ast.OrdChoice("EXPRESSION", nil, boolCondition, condition)

	return &Lexer{ast, expression}, nil
}
