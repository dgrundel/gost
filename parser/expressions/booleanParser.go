package expressions

import (
	"bytes"
	"fmt"
)

type parser struct {
	tokens []string
	pos    int
}

// ParseBooleanExpression creates a new BooleanExpression by parsing the input string
// using a Pratt parser implementation.
// order of operations:
// 1. parenthesized expressions
// 2. logical NOT
// 3. comparison operators
// 4. logical AND
// 5. logical OR
func ParseBooleanExpression(s string) (BooleanExpression, error) {
	p := &parser{
		tokens: tokenize(s),
	}

	if len(p.tokens) == 0 {
		return nil, fmt.Errorf("empty expression")
	}

	return p.parseWithPrecedence(0)
}

// Get precedence level for operators
func precedence(op BooleanOperator) int {
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
func tokenize(s string) []string {
	var tokens []string
	var token bytes.Buffer
	runes := []rune(s)

	for i := 0; i < len(runes); i++ {
		c := runes[i]

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
			if i+1 < len(runes) {
				next := runes[i+1]
				if (c == '=' && next == '=') ||
					(c == '!' && next == '=') ||
					(c == '>' && next == '=') ||
					(c == '<' && next == '=') ||
					(c == '&' && next == '&') ||
					(c == '|' && next == '|') {
					tokens = append(tokens, string(c)+string(next))
					i++ // Skip the next rune
					continue
				}
			}
			tokens = append(tokens, string(c))
		default:
			token.WriteRune(c)
		}
	}

	if token.Len() > 0 {
		tokens = append(tokens, token.String())
	}

	return tokens
}

// Parse with given precedence level
func (p *parser) parseWithPrecedence(precedenceLevel int) (BooleanExpression, error) {
	var left BooleanExpression

	// Handle prefix operators and literals
	token := p.tokens[p.pos]
	p.pos++

	switch token {
	case "(":
		expr, err := p.parseWithPrecedence(0)
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
		right, err := p.parseWithPrecedence(precedence(LogicalNot))
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

		right, err := p.parseWithPrecedence(nextPrecedence)
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
