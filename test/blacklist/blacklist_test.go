package blacklist

import (
	"goaway/internal/blacklist"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type Blacklist struct{}

func TestExtractDomains(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
		hasError bool
	}{
		{
			name: "ignore comments and empty lines",
			input: `
			# This and the following lines shall be ignored
			

			`,
			expected: nil,
			hasError: true,
		},
		{
			name:     "extract valid domain from 0.0.0.0 line",
			input:    "0.0.0.0 example.com",
			expected: []string{"example.com"},
		},
		{
			name: "ignore localhost and similar",
			input: `
			0.0.0.0 localhost
			0.0.0.0 localhost.localdomain
			0.0.0.0 broadcasthost
			0.0.0.0 local
			`,
			expected: nil,
			hasError: true,
		},
		{
			name: "mixed valid and ignored domains",
			input: `
			# Comment
			0.0.0.0 example.com
			0.0.0.0 localhost
			0.0.0.0 tracker.example.org
			127.0.0.1 another.com
			`,
			expected: []string{"example.com", "tracker.example.org", "another.com"},
		},
		{
			name:     "single domain no IP",
			input:    "suspiciousdomain.com",
			expected: []string{"suspiciousdomain.com"},
		},
	}

	b := &blacklist.Blacklist{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := b.ExtractDomains(strings.NewReader(tt.input))

			if tt.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, result)
			}
		})
	}
}
