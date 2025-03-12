package expressions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseBooleanExpression(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      string
		wantTypes map[string]ExpressionType
		wantErr   bool
	}{
		{
			name:  "simple equality",
			input: "a == b",
			want:  "a == b",
		},
		{
			name:  "simple equality with type declaration",
			input: "a:string == b:string",
			want:  "a:string == b:string",
			wantTypes: map[string]ExpressionType{
				"a": NewExpressionType(ExpressionBaseTypeString, "", ExpressionBaseTypeString),
				"b": NewExpressionType(ExpressionBaseTypeString, "", ExpressionBaseTypeString),
			},
		},
		{
			name:  "simple inequality",
			input: "x != y",
			want:  "x != y",
		},
		{
			name:  "greater than",
			input: "count > 5",
			want:  "count > 5",
		},
		{
			name:  "less than or equal",
			input: "value <= 10",
			want:  "value <= 10",
		},
		{
			name:  "logical AND",
			input: "a == b && c != d",
			want:  "a == b && c != d",
		},
		{
			name:  "logical OR",
			input: "x > 5 || y < 3",
			want:  "x > 5 || y < 3",
		},
		{
			name:  "logical NOT",
			input: "!isValid",
			want:  "! isValid",
		},
		{
			name:  "logical NOT with type declaration",
			input: "!isValid:bool",
			want:  "! isValid:bool",
			wantTypes: map[string]ExpressionType{
				"isValid": NewExpressionType(ExpressionBaseTypeBool, "", ExpressionBaseTypeBool),
			},
		},
		{
			name:  "parenthesized expression",
			input: "(a == b)",
			want:  "(a == b)",
		},
		{
			name:  "complex expression with precedence",
			input: "a == b && (c > d || e != f)",
			want:  "a == b && (c > d || e != f)",
		},
		{
			name:  "multiple logical operators",
			input: "x > 5 && y < 10 || z == 0",
			want:  "x > 5 && y < 10 || z == 0",
		},
		{
			name:  "nested parentheses",
			input: "((a == b) && c) || d",
			want:  "((a == b) && c) || d",
		},
		{
			name:  "nested parentheses with type declaration",
			input: "((a:int == b:int) && c:bool) || d:bool",
			want:  "((a:int == b:int) && c:bool) || d:bool",
			wantTypes: map[string]ExpressionType{
				"a": NewExpressionType(ExpressionBaseTypeInt, "", ExpressionBaseTypeInt),
				"b": NewExpressionType(ExpressionBaseTypeInt, "", ExpressionBaseTypeInt),
				"c": NewExpressionType(ExpressionBaseTypeBool, "", ExpressionBaseTypeBool),
				"d": NewExpressionType(ExpressionBaseTypeBool, "", ExpressionBaseTypeBool),
			},
		},
		{
			name:    "empty expression",
			input:   "",
			wantErr: true,
		},
		{
			name:    "missing closing parenthesis",
			input:   "(a == b",
			wantErr: true,
		},
		{
			name:    "invalid type",
			input:   "a:invalid == b",
			wantErr: true,
		},
		{
			name:    "missing type after colon",
			input:   "a: == b",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, types, err := ParseBooleanExpression(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			if got != nil {
				assert.Equal(t, tt.want, got.String())
			}
			if tt.wantTypes != nil {
				assert.Equal(t, tt.wantTypes, types)
			}
		})
	}
}
