package ast

import (
	"bytes"

	"github.com/JosueMolinaMorales/monkeylang/internal/token"
)

// Node represents a node in the AST
type Node interface {
	// TokenLiteral returns the literal value of the token associated with the node
	TokenLiteral() string
	// String returns a string representation of the node
	String() string
}

// Statement represents a statement node in the AST
type Statement interface {
	// Node represents a node in the AST
	Node
	// statementNode represents a statement node in the AST
	statementNode()
}

// Expression represents an expression node in the AST
type Expression interface {
	// Node represents a node in the AST
	Node
	// expressionNode represents an expression node in the AST
	expressionNode()
}

// LetStatement represents a let statment node in the AST
type LetStatement struct {
	Token token.Token // the token.LET Token
	Name  *Identifier
	Value Expression
}

func (ls *LetStatement) statementNode() {}

// TokenLiteral returns the literal value of the token associated with the node
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }

func (ls *LetStatement) String() string {
	var out bytes.Buffer

	out.WriteString(ls.TokenLiteral() + " ")
	out.WriteString(ls.Name.String())
	out.WriteString(" = ")

	if ls.Value != nil {
		out.WriteString(ls.Value.String())
	}

	out.WriteString(";")

	return out.String()
}

// Identifier represents an identifier node in the AST
type Identifier struct {
	Token token.Token // the token.IDENT Token
	Value string
}

func (i *Identifier) expressionNode() {}

// TokenLiteral returns the literal value of the token associated with the node
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

// Program represents a program node in the AST
type Program struct {
	Statements []Statement
}

func (p *Program) String() string {
	var out bytes.Buffer

	for _, s := range p.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}

// TokenLiteral returns the literal value of the token associated with the node
func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

// ReturnStatement represents a return statement node in the AST
type ReturnStatement struct {
	Token       token.Token // the 'return' Token
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode() {}

// TokenLiteral returns the literal value of the token associated with the node
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }

func (rs *ReturnStatement) String() string {
	var out bytes.Buffer

	out.WriteString(rs.TokenLiteral() + " ")

	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String())
	}

	out.WriteString(";")

	return out.String()
}

// ExpressionStatement represents an expression statement node in the AST
type ExpressionStatement struct {
	Token      token.Token // the first token of the expression
	Expression Expression
}

func (es *ExpressionStatement) statementNode() {}

// TokenLiteral returns the literal value of the token associated with the node
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }

func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}

	return ""
}
