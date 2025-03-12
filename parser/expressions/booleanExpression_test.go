package expressions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseBooleanExpression(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "simple equality",
			input: "a == b",
			want:  "a == b",
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
			name:    "empty expression",
			input:   "",
			wantErr: true,
		},
		{
			name:    "missing closing parenthesis",
			input:   "(a == b",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewBooleanExpression(tt.input)
			assert.Equal(t, tt.wantErr, err != nil)
			if err != nil {
				return
			}
			assert.Equal(t, tt.want, got.String())
		})
	}
}
