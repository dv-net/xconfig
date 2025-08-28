package utils

import (
	"reflect"
	"testing"
)

func TestSplitNameByWords(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "single lowercase word",
			input:    "word",
			expected: []string{"word"},
		},
		{
			name:     "single uppercase word",
			input:    "WORD",
			expected: []string{"WORD"},
		},
		{
			name:     "camel case",
			input:    "camelCase",
			expected: []string{"camel", "Case"},
		},
		{
			name:     "pascal case",
			input:    "PascalCase",
			expected: []string{"Pascal", "Case"},
		},
		{
			name:     "snake case",
			input:    "snake_case",
			expected: []string{"snake", "_", "case"},
		},
		{
			name:     "mixed case with numbers",
			input:    "Mixed123Case",
			expected: []string{"Mixed", "123", "Case"},
		},
		{
			name:     "acronyms",
			input:    "PDFLoader",
			expected: []string{"PDF", "Loader"},
		},
		{
			name:     "multiple uppercase letters followed by lowercase",
			input:    "HTTPRequest",
			expected: []string{"HTTP", "Request"},
		},
		{
			name:     "special characters",
			input:    "special-chars@here",
			expected: []string{"special", "-", "chars", "@", "here"},
		},
		{
			name:     "mixed case with special chars and numbers",
			input:    "User123_ID-Info",
			expected: []string{"User", "123", "_", "ID", "-", "Info"},
		},
		{
			name:     "word and number mix",
			input:    "S3",
			expected: []string{"S3"},
		},
		{
			name:     "word number mix and uppercased letters",
			input:    "BaseUrlS3API",
			expected: []string{"Base", "Url", "S3", "API"},
		},
		{
			name:     "word, dot and uppercased letters",
			input:    "BaseURL.API",
			expected: []string{"Base", "URL", "API"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SplitNameByWords(tt.input)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("SplitNameByWords(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}
