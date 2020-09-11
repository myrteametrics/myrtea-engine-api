package lexer

import (
	"testing"

	parsec "github.com/prataprc/goparsec"
)

func TestReplaceGlobal(t *testing.T) {
	modelEntities := []string{"order.zipcode", "order"}
	l, err := New(modelEntities)
	if err != nil {
		t.Error(err)
	}
	reverse := ReplaceGlobals(l)
	if L() == nil {
		t.Error("Global lexer is nil")
	}
	reverse()
	if L() != nil {
		t.Error("Global lexer is not nil after reverse")
	}
}

func TestLexerConditionExists(t *testing.T) {
	modelEntities := []string{"order.zipcode", "order"}
	conditionStr := "EXISTS order.zipcode"

	l, err := New(modelEntities)
	if err != nil {
		t.Error(err)
	}
	nodes, _ := l.Ast.Parsewith(l.Parser, parsec.NewScanner([]byte(conditionStr)))
	if nodes == nil {
		t.Error("root is nil")
	}
	l.Ast.Prettyprint()
	if nodes.GetName() != "EXISTS" {
		t.Error("Invalid parsing")
	}
	if len(nodes.GetChildren()) != 2 {
		t.Error("Invalid childrens")
	}
	if nodes.GetChildren()[0].GetName() != "EXISTS" {
		t.Error("Invalid childrens[0]")
	}
	if nodes.GetChildren()[1].GetName() != "ENTITY" || nodes.GetChildren()[1].GetValue() != "order.zipcode" {
		t.Error("Invalid childrens[1]")
	}
}

func TestLexerConditionBetween(t *testing.T) {
	modelEntities := []string{"order.in_timestamp", "order"}
	conditionStr := "begin < order.in_timestamp < now"

	l, err := New(modelEntities)
	if err != nil {
		t.Error(err)
	}
	nodes, _ := l.Ast.Parsewith(l.Parser, parsec.NewScanner([]byte(conditionStr)))
	if nodes == nil {
		t.Error("root is nil")
	}
	l.Ast.Prettyprint()
	if nodes.GetName() != "BETWEEN" {
		t.Error("Invalid parsing")
	}
	if len(nodes.GetChildren()) != 5 {
		t.Error("Invalid childrens")
	}
	if nodes.GetChildren()[0].GetName() != "VALUE" || nodes.GetChildren()[0].GetValue() != "begin" {
		t.Error("Invalid childrens[0]")
	}
	if nodes.GetChildren()[1].GetName() != "LESS_THAN" {
		t.Error("Invalid childrens[1]")
	}
	if nodes.GetChildren()[2].GetName() != "ENTITY" || nodes.GetChildren()[2].GetValue() != "order.in_timestamp" {
		t.Error("Invalid childrens[2]")
	}
	if nodes.GetChildren()[3].GetName() != "LESS_THAN" {
		t.Error("Invalid childrens[3]")
	}
	if nodes.GetChildren()[4].GetName() != "VALUE" || nodes.GetChildren()[4].GetValue() != "now" {
		t.Error("Invalid childrens[4]")
	}
}

func TestLexerConditionCompareEqual(t *testing.T) {
	modelEntities := []string{"order.client", "order"}
	conditionStr := "order.client = montlouis"

	l, err := New(modelEntities)
	if err != nil {
		t.Error(err)
	}
	nodes, _ := l.Ast.Parsewith(l.Parser, parsec.NewScanner([]byte(conditionStr)))
	if nodes == nil {
		t.Error("root is nil")
	}
	l.Ast.Prettyprint()
	if nodes.GetName() != "COMPARE" {
		t.Error("Invalid parsing")
	}
	if len(nodes.GetChildren()) != 3 {
		t.Error("Invalid childrens")
	}
	if nodes.GetChildren()[0].GetName() != "ENTITY" || nodes.GetChildren()[0].GetValue() != "order.client" {
		t.Error("Invalid childrens[0]")
	}
	if nodes.GetChildren()[1].GetName() != "EQUALS" {
		t.Error("Invalid childrens[1]")
	}
	if nodes.GetChildren()[2].GetName() != "VALUE" || nodes.GetChildren()[2].GetValue() != "montlouis" {
		t.Error("Invalid childrens[2]")
	}
}

func TestLexerConditionScript(t *testing.T) {
	modelEntities := []string{}
	conditionStr := "${doc.dpdex_done_date.value.millis > doc.dpdfiledist_date.value.millis}"

	l, err := New(modelEntities)
	if err != nil {
		t.Error(err)
	}
	nodes, _ := l.Ast.Parsewith(l.Parser, parsec.NewScanner([]byte(conditionStr)))
	if nodes == nil {
		t.Error("root is nil")
	}
	l.Ast.Prettyprint()
	if nodes.GetName() != "SCRIPT" {
		t.Error("Invalid parsing")
	}
	if len(nodes.GetChildren()) != 3 {
		t.Error("Invalid childrens")
	}
	if nodes.GetChildren()[0].GetName() != "OPEN_SCRIPT" {
		t.Error("Invalid childrens[0]")
	}
	if nodes.GetChildren()[1].GetName() != "SCRIPT_CONTENT" ||
		nodes.GetChildren()[1].GetValue() != "doc.dpdex_done_date.value.millis > doc.dpdfiledist_date.value.millis" {
		t.Error("Invalid childrens[1]")
	}
	if nodes.GetChildren()[2].GetName() != "CLOSE_SCRIPT" {
		t.Error("Invalid childrens[2]")
	}
}

func TestLexerConditionScript2(t *testing.T) {
	modelEntities := []string{}
	conditionStr := "${(doc.dpdex_done_date.value.millis - doc.dpdfiledist_date.value.millis) / 1000 > 0}"

	l, err := New(modelEntities)
	if err != nil {
		t.Error(err)
	}
	nodes, _ := l.Ast.Parsewith(l.Parser, parsec.NewScanner([]byte(conditionStr)))
	if nodes == nil {
		t.Error("root is nil")
	}
	l.Ast.Prettyprint()
	if nodes.GetName() != "SCRIPT" {
		t.Error("Invalid parsing")
	}
	if len(nodes.GetChildren()) != 3 {
		t.Error("Invalid childrens")
	}
	if nodes.GetChildren()[0].GetName() != "OPEN_SCRIPT" {
		t.Error("Invalid childrens[0]")
	}
	if nodes.GetChildren()[1].GetName() != "SCRIPT_CONTENT" ||
		nodes.GetChildren()[1].GetValue() != "(doc.dpdex_done_date.value.millis - doc.dpdfiledist_date.value.millis) / 1000 > 0" {
		t.Error("Invalid childrens[1]")
	}
	if nodes.GetChildren()[2].GetName() != "CLOSE_SCRIPT" {
		t.Error("Invalid childrens[2]")
	}
}
