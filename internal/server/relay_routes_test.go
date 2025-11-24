package server

import "testing"

func TestDetermineUpstreamPath(t *testing.T) {
	cases := []struct {
		model    string
		expected string
	}{
		{"gemini-1.5-pro", "/v1beta/models/gemini-1.5-pro:generateContent"},
		{"claude-3", "/v1/messages"},
		{"gpt-4o", "/v1/chat/completions"},
	}

	for _, tc := range cases {
		if got := determineUpstreamPath(tc.model); got != tc.expected {
			t.Fatalf("model %s: expected %s, got %s", tc.model, tc.expected, got)
		}
	}
}

func TestExtractGeminiModelName(t *testing.T) {
	cases := []struct {
		path     string
		expected string
	}{
		{"models/gemini-1.5-flash:generateContent", "gemini-1.5-flash"},
		{"/v1beta/models/gemini-pro:generateContent", "gemini-pro"},
		{"/models/gemini-ultra", "gemini-ultra"},
		{"", ""},
	}

	for _, tc := range cases {
		if got := extractGeminiModelName(tc.path); got != tc.expected {
			t.Fatalf("path %s: expected %s, got %s", tc.path, tc.expected, got)
		}
	}
}
