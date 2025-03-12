package typescript

import (
	"bytes"
	"gost/parser"
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	tests := []struct {
		name     string
		template string
	}{
		{
			name: "loop block",
			template: `<ul>
				{for i, item in items: string[]}
					<li data-index={i}>{item}</li>
				{/for}
			</ul>`,
		},
		{
			name: "conditional block with else-if and else",
			template: `<div>
				{if qty: int == 1}
					You have {qty} item.
				{else if qty > 1000}
					You have way too many items.
				{else}
					You have {qty} items.
				{/if}
			</div>`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			r := strings.NewReader(tt.template)
			document, err := parser.Parse(r)
			if err != nil {
				t.Fatalf("failed to parse template: %v", err)
			}

			var buf bytes.Buffer
			err = Generate(document, &buf)
			if err != nil {
				t.Fatalf("failed to generate template: %v", err)
			}

			t.Log("Parsed template:", document.String())
			t.Log("Generated typescript:", buf.String())
		})
	}
}
