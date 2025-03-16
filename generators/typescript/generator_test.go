package typescript

import (
	"bytes"
	"gost/parser"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerate(t *testing.T) {
	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name: "loop block",
			template: `<ul>
				{for i, item in items: string[]}
					<li data-index={i}>{item}</li>
				{/for}
			</ul>`,
			expected: strings.Join([]string{
				"const encoder = document.createElement('div');",
				"const htmlEncode = (value: string) => {",
				"	encoder.textContent = value;",
				"	return encoder.innerHTML;",
				"};",
				"export interface model {",
				"	items: string[];",
				"}",
				"export const render = ({items}: model) => (`<ul>",
				"				${[...(Array.isArray(items) ? items.entries() : Object.entries(items))].map(([i, item]) => (`",
				"					<li data-index=\"${htmlEncode(`${i}`)}\">${item}</li>",
				"				`)).join('')}",
				"			</ul>`);",
			}, "\n"),
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
			expected: strings.Join([]string{
				"const encoder = document.createElement('div');",
				"const htmlEncode = (value: string) => {",
				"	encoder.textContent = value;",
				"	return encoder.innerHTML;",
				"};",
				"export interface model {",
				"	qty: number;",
				"}",
				"export const render = ({qty}: model) => (`<div>",
				"				${(qty == 1) && (`",
				"					You have ${qty} item.",
				"				`) || (qty > 1000) && (`",
				"					You have way too many items.",
				"				`) || (`",
				"					You have ${qty} items.",
				"				`) || ''}",
				"			</div>`);",
			}, "\n"),
		}, {
			name: "spread attribute",
			template: `<div>
				<img {...attrs: map[string,string]}>
			</div>`,
			expected: strings.Join([]string{
				"const encoder = document.createElement('div');",
				"const htmlEncode = (value: string) => {",
				"	encoder.textContent = value;",
				"	return encoder.innerHTML;",
				"};",
				"export interface model {",
				"	attrs: Record<string,string>;",
				"}",
				"export const render = ({attrs}: model) => (`<div>",
				"				<img ${[...Object.entries(attrs)].map(([k, v]) => (`${k}=\"${htmlEncode(v)}\"`)).join(' ')}>",
				"			</div>`);",
			}, "\n"),
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

			// t.Log("Parsed template:", document.String())
			t.Log("Generated typescript:", buf.String())
			assert.Equal(t, tt.expected, buf.String())
		})
	}
}
