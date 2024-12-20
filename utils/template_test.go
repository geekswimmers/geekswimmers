package utils

import (
	"html/template"
	"testing"
)

func TestTitle(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "lowercase",
			input:    "freestyle",
			expected: "Freestyle",
		},
		{
			name:     "uppercase",
			input:    "FREESTYLE",
			expected: "Freestyle",
		},
		{
			name:     "mixed case",
			input:    "FreEsTyLe",
			expected: "Freestyle",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "underscore",
			input:    "FREESTYLE_RELAY",
			expected: "Freestyle Relay",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := Title(tc.input)
			if actual != tc.expected {
				t.Errorf("Title(%q) = %q, want %q", tc.input, actual, tc.expected)
			}
		})
	}
}

func TestToHTML(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected template.HTML
	}{
		{
			name:  "basic markdown",
			input: "# Heading\n\nParagraph text",
			expected: template.HTML(`<h1>Heading</h1>
<p>Paragraph text</p>
`),
		},
		{
			name:     "empty input",
			input:    "",
			expected: template.HTML(""),
		},
		{
			name:  "markdown with link",
			input: "[Example](https://example.com)",
			expected: template.HTML(`<p><a href="https://example.com" target="_blank" rel="noopener noreferrer">Example</a></p>
`),
		},
		{
			name:  "markdown with image",
			input: "![Image](image.jpg)",
			expected: template.HTML(`<p><img src="image.jpg" class="img-fluid" alt="Image"></p>
`),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := MarkdownToHTML(tc.input)
			if actual != tc.expected {
				t.Errorf("MarkdownToHTML(%q) = %q, want %q", tc.input, actual, tc.expected)
			}
		})
	}
}
