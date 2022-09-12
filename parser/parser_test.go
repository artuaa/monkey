package parser

import (
	"interpreter/ast"
	"interpreter/lexer"
	"testing"
)

func TestLetStatements(t *testing.T) {
	input := `let x = 5;
let y = 10;
let foobar = 838383;`

	l := lexer.New(input)

	p := New(l)

	program := p.ParseProgram()

	checkParserErrors(t, p)

	if program == nil {
		t.Fatalf("parse program returned nil")
	}

	if len(program.Statements) != 3 {
		t.Fatalf("program statementes doesn't contain 3 statements. got=%d", len(program.Statements))
	}

	tests := []struct {
		expectedIdentifier string
	}{{"x"}, {"y"}, {"foobar"}}

	for i, tt := range tests {
		stmt := program.Statements[i]
		if !testLetStatement(t, stmt, tt.expectedIdentifier) {
			return
		}
	}
}

func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()

	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d erros", len(errors))

	for _, msg := range errors {
		t.Errorf("parse error: %s", msg)
	}
	t.FailNow()
}

func testLetStatement(t *testing.T, stmt ast.Statement, name string) bool {
	if l := stmt.TokenLiteral(); l != "let" {
		t.Errorf("token literal not 'let'. got=%q", l)
	}

	letStmt, ok := stmt.(*ast.LetStatement)
	if !ok {
		t.Errorf("s not let statement. got=%T", stmt)
		return false
	}

	if letStmt.Name.Value != name {
		t.Errorf("statement value not %s. got=%s", name, letStmt.Name.Value)
		return false
	}

	if letStmt.Name.TokenLiteral() != name {
		t.Errorf("statment name not %s. got=%s", name, letStmt.Name)
		return false
	}
	return true
}

func TestReturnStatements(t *testing.T) {
	input := `
return 5;
return 10;
return 993322;`

	l := lexer.New(input)

	p := New(l)

	program := p.ParseProgram()

	checkParserErrors(t, p)

	if program == nil {
		t.Fatalf("parse program returned nil")
	}

	if len(program.Statements) != 3 {
		t.Fatalf("program statementes doesn't contain 3 statements. got=%d", len(program.Statements))
	}

	for _, stmt := range program.Statements {
		returnStmt, ok := stmt.(*ast.ReturnStatement)

		if !ok {
			t.Errorf("smnt not return statement. got=%T", stmt)
			continue
		}

		if l := returnStmt.TokenLiteral(); l != "return" {
			t.Errorf("statement literal not 'return'. got %q", l)
		}
	}
}
func TestIdentifierExpression(t *testing.T) {
	input := "foobar;"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	if len(program.Statements) != 1 {
		t.Fatalf("program has not enough statements. got=%d",
			len(program.Statements))
	}
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}
	ident, ok := stmt.Expression.(*ast.Identifier)
	if !ok {
		t.Fatalf("exp not *ast.Identifier. got=%T", stmt.Expression)
	}
	if ident.Value != "foobar" {
		t.Errorf("ident.Value not %s. got=%s", "foobar", ident.Value)
	}
	if ident.TokenLiteral() != "foobar" {
		t.Errorf("ident.TokenLiteral not %s. got=%s", "foobar",
			ident.TokenLiteral())
	}
}
