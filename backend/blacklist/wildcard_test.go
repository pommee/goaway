package blacklist

import (
	"testing"
)

func TestMatchesWildcard(t *testing.T) {
	tests := []struct {
		name     string
		domain   string
		pattern  string
		expected bool
	}{
		{
			name:     "simple wildcard match",
			domain:   "test.example.com",
			pattern:  "*.example.com",
			expected: true,
		},
		{
			name:     "simple wildcard no match different domain",
			domain:   "test.other.com",
			pattern:  "*.example.com",
			expected: false,
		},
		{
			name:     "wildcard with trailing dot",
			domain:   "test.example.com",
			pattern:  "*.example.com.",
			expected: true,
		},
		{
			name:     "domain with trailing dot",
			domain:   "test.example.com.",
			pattern:  "*.example.com",
			expected: true,
		},
		{
			name:     "both with trailing dots",
			domain:   "test.example.com.",
			pattern:  "*.example.com.",
			expected: true,
		},
		{
			name:     "multi-level subdomain matches single wildcard",
			domain:   "a.b.example.com",
			pattern:  "*.example.com",
			expected: true,
		},
		{
			name:     "www subdomain matches",
			domain:   "www.example.com",
			pattern:  "*.example.com",
			expected: true,
		},
		{
			name:     "no match when domain is the base",
			domain:   "example.com",
			pattern:  "*.example.com",
			expected: false,
		},
		{
			name:     "partial domain no match",
			domain:   "test.example.com.uk",
			pattern:  "*.example.com",
			expected: false,
		},
		{
			name:     "testing.some.domain matches *.some.domain",
			domain:   "testing.some.domain",
			pattern:  "*.some.domain",
			expected: true,
		},
		{
			name:     "deep subdomain matches single wildcard",
			domain:   "a.b.c.d.example.com",
			pattern:  "*.example.com",
			expected: true,
		},
		{
			name:     "api.github.com as pattern matches api.github.com domain",
			domain:   "api.github.com",
			pattern:  "api.github.com",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesWildcard(tt.domain, tt.pattern)
			if result != tt.expected {
				t.Errorf("matchesWildcard(%q, %q) = %v, want %v", tt.domain, tt.pattern, result, tt.expected)
			}
		})
	}
}

func TestIsValidDomainOrWildcard(t *testing.T) {
	tests := []struct {
		name     string
		domain   string
		expected bool
	}{
		{
			name:     "valid fqdn",
			domain:   "example.com",
			expected: true,
		},
		{
			name:     "valid fqdn with subdomain",
			domain:   "sub.example.com",
			expected: true,
		},
		{
			name:     "valid wildcard",
			domain:   "*.example.com",
			expected: true,
		},
		{
			name:     "valid wildcard with trailing dot",
			domain:   "*.example.com.",
			expected: true,
		},
		{
			name:     "valid wildcard with subdomain",
			domain:   "*.sub.example.com",
			expected: true,
		},
		{
			name:     "no dot",
			domain:   "localhost",
			expected: false,
		},
		{
			name:     "wildcard with only TLD",
			domain:   "*.com",
			expected: false,
		},
		{
			name:     "double wildcard",
			domain:   "*.*.example.com",
			expected: false,
		},
		{
			name:     "wildcard in middle",
			domain:   "sub.*.example.com",
			expected: false,
		},
		{
			name:     "only wildcard",
			domain:   "*",
			expected: false,
		},
		{
			name:     "empty string",
			domain:   "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidDomainOrWildcard(tt.domain)
			if result != tt.expected {
				t.Errorf("isValidDomainOrWildcard(%q) = %v, want %v", tt.domain, result, tt.expected)
			}
		})
	}
}
