package expressions

import (
	"bytes"
	"fmt"
)

// Supported boolean operators:
type BooleanOperator string

const (
	Equal              BooleanOperator = "=="
	NotEqual           BooleanOperator = "!="
	GreaterThan        BooleanOperator = ">"
	LessThan           BooleanOperator = "<"
	GreaterThanOrEqual BooleanOperator = ">="
	LessThanOrEqual    BooleanOperator = "<="
	LogicalAnd         BooleanOperator = "&&"
	LogicalOr          BooleanOperator = "||"
	LogicalNot         BooleanOperator = "!"
)

var _booleanOperators = map[string]BooleanOperator{
	"==": Equal,
	"!=": NotEqual,
	">":  GreaterThan,
	"<":  LessThan,
	">=": GreaterThanOrEqual,
	"<=": LessThanOrEqual,
	"&&": LogicalAnd,
	"||": LogicalOr,
	"!":  LogicalNot,
}

// Also supported:
// parenthesized expressions
// && and || are short-circuiting

type BooleanExpression interface {
	Left() BooleanExpression
	Right() BooleanExpression
	Operator() BooleanOperator
	Parentheses() bool
	Literal() string
	String() string
}

type booleanExpression struct {
	left        BooleanExpression
	right       BooleanExpression
	operator    BooleanOperator
	literal     string
	parentheses bool
}

// order of operations:
// 1. parenthesized expressions
// 2. logical NOT
// 3. comparison operators
// 4. logical AND
// 5. logical OR
func ParseBooleanExpression(s string) (BooleanExpression, error) {
	// Pratt parser implementation for boolean expressions
	type parser struct {
		tokens []string
		pos    int
	}

	// Get precedence level for operators
	precedence := func(op BooleanOperator) int {
		switch op {
		case LogicalOr:
			return 1
		case LogicalAnd:
			return 2
		case Equal, NotEqual, GreaterThan, LessThan, GreaterThanOrEqual, LessThanOrEqual:
			return 3
		case LogicalNot:
			return 4
		default:
			return 0
		}
	}

	// Tokenize input string
	tokenize := func(s string) []string {
		var tokens []string
		var token bytes.Buffer

		for i := 0; i < len(s); i++ {
			c := s[i]

			switch c {
			case ' ', '\t', '\n', '\r':
				if token.Len() > 0 {
					tokens = append(tokens, token.String())
					token.Reset()
				}
			case '(', ')':
				if token.Len() > 0 {
					tokens = append(tokens, token.String())
					token.Reset()
				}
				tokens = append(tokens, string(c))
			case '=', '!', '>', '<', '&', '|':
				if token.Len() > 0 {
					tokens = append(tokens, token.String())
					token.Reset()
				}

				// Handle two-character operators
				if i+1 < len(s) {
					next := s[i+1]
					if (c == '=' && next == '=') ||
						(c == '!' && next == '=') ||
						(c == '>' && next == '=') ||
						(c == '<' && next == '=') ||
						(c == '&' && next == '&') ||
						(c == '|' && next == '|') {
						tokens = append(tokens, string(c)+string(next))
						i++
						continue
					}
				}
				tokens = append(tokens, string(c))
			default:
				token.WriteByte(c)
			}
		}

		if token.Len() > 0 {
			tokens = append(tokens, token.String())
		}

		return tokens
	}

	p := &parser{
		tokens: tokenize(s),
	}

	// Parse with given precedence level
	var parseWithPrecedence func(precedenceLevel int) (BooleanExpression, error)
	parseWithPrecedence = func(precedenceLevel int) (BooleanExpression, error) {
		var left BooleanExpression

		// Handle prefix operators and literals
		token := p.tokens[p.pos]
		p.pos++

		switch token {
		case "(":
			expr, err := parseWithPrecedence(0)
			if err != nil {
				return nil, err
			}
			if p.pos >= len(p.tokens) || p.tokens[p.pos] != ")" {
				return nil, fmt.Errorf("missing closing parenthesis")
			}
			p.pos++
			left = &booleanExpression{
				left:        expr.Left(),
				right:       expr.Right(),
				operator:    expr.Operator(),
				literal:     expr.Literal(),
				parentheses: true,
			}
		case "!":
			right, err := parseWithPrecedence(precedence(LogicalNot))
			if err != nil {
				return nil, err
			}
			left = &booleanExpression{
				operator: LogicalNot,
				right:    right,
			}
		default:
			left = &booleanExpression{literal: token}
		}

		// Handle infix operators
		for p.pos < len(p.tokens) {
			token = p.tokens[p.pos]

			op, isOp := _booleanOperators[token]
			if !isOp {
				break
			}

			nextPrecedence := precedence(op)
			if nextPrecedence <= precedenceLevel {
				break
			}

			p.pos++

			right, err := parseWithPrecedence(nextPrecedence)
			if err != nil {
				return nil, err
			}

			left = &booleanExpression{
				left:     left,
				right:    right,
				operator: op,
			}
		}

		return left, nil
	}

	if len(p.tokens) == 0 {
		return nil, fmt.Errorf("empty expression")
	}

	return parseWithPrecedence(0)

}

func (b *booleanExpression) Left() BooleanExpression {
	return b.left
}

func (b *booleanExpression) Right() BooleanExpression {
	return b.right
}

func (b *booleanExpression) Operator() BooleanOperator {
	return b.operator
}

func (b *booleanExpression) Parentheses() bool {
	return b.parentheses
}

func (b *booleanExpression) Literal() string {
	return b.literal
}

func (b *booleanExpression) String() string {
	if b.literal != "" {
		return b.literal
	}

	var buf bytes.Buffer
	if b.parentheses {
		buf.WriteByte('(')
	}
	if b.left != nil {
		buf.WriteString(b.left.String())
		buf.WriteByte(' ')
	}
	buf.WriteString(string(b.operator))
	if b.right != nil {
		buf.WriteByte(' ')
		buf.WriteString(b.right.String())
	}
	if b.parentheses {
		buf.WriteByte(')')
	}
	return buf.String()
}
