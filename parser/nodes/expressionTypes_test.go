package nodes

import (
	"testing"
)

func TestParseExpressionType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		want     ExpressionType
		wantBool bool
	}{
		{
			name:     "simple string type",
			input:    "string",
			want:     NewExpressionType(ExpressionBaseTypeString, "", ExpressionBaseTypeString),
			wantBool: true,
		},
		{
			name:     "simple int type",
			input:    "int",
			want:     NewExpressionType(ExpressionBaseTypeInt, "", ExpressionBaseTypeInt),
			wantBool: true,
		},
		{
			name:     "array of strings",
			input:    "string[]",
			want:     NewExpressionType(ExpressionBaseTypeArray, ExpressionBaseTypeInt, ExpressionBaseTypeString),
			wantBool: true,
		},
		{
			name:     "array of ints",
			input:    "int[]",
			want:     NewExpressionType(ExpressionBaseTypeArray, ExpressionBaseTypeInt, ExpressionBaseTypeInt),
			wantBool: true,
		},
		{
			name:     "map with string key and int value",
			input:    "map[string, int]",
			want:     NewExpressionType(ExpressionBaseTypeMap, ExpressionBaseTypeString, ExpressionBaseTypeInt),
			wantBool: true,
		},
		{
			name:     "map with int key and bool value",
			input:    "map[int, bool]",
			want:     NewExpressionType(ExpressionBaseTypeMap, ExpressionBaseTypeInt, ExpressionBaseTypeBool),
			wantBool: true,
		},
		{
			name:     "invalid type",
			input:    "invalid",
			want:     nil,
			wantBool: false,
		},
		{
			name:     "invalid array syntax",
			input:    "string[",
			want:     nil,
			wantBool: false,
		},
		{
			name:     "invalid map syntax",
			input:    "map[string]",
			want:     nil,
			wantBool: false,
		},
		{
			name:     "map with invalid key type",
			input:    "map[invalid, int]",
			want:     nil,
			wantBool: false,
		},
		{
			name:     "map with invalid value type",
			input:    "map[int, invalid]",
			want:     nil,
			wantBool: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ParseExpressionType(tt.input)
			if ok != tt.wantBool {
				t.Errorf("ParseExpressionType() ok = %v, want %v", ok, tt.wantBool)
				return
			}
			if !tt.wantBool {
				return
			}
			if got.BaseType() != tt.want.BaseType() {
				t.Errorf("ParseExpressionType() baseType = %v, want %v", got.BaseType(), tt.want.BaseType())
			}
			if got.KeyType() != tt.want.KeyType() {
				t.Errorf("ParseExpressionType() keyType = %v, want %v", got.KeyType(), tt.want.KeyType())
			}
			if got.ValueType() != tt.want.ValueType() {
				t.Errorf("ParseExpressionType() valueType = %v, want %v", got.ValueType(), tt.want.ValueType())
			}
		})
	}
}
