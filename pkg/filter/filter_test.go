package filter

import (
	"reflect"
	"testing"
)

func TestFilterContexts(t *testing.T) {
	tests := []struct {
		name     string
		contexts []string
		include  []string
		exclude  []string
		want     []string
	}{
		{
			name:     "No filters",
			contexts: []string{"dev", "staging", "prod"},
			include:  []string{},
			exclude:  []string{},
			want:     []string{"dev", "staging", "prod"},
		},
		{
			name:     "Include filter only",
			contexts: []string{"dev", "staging", "prod", "prod-eu", "prod-us"},
			include:  []string{"prod"},
			exclude:  []string{},
			want:     []string{"prod", "prod-eu", "prod-us"},
		},
		{
			name:     "Exclude filter only",
			contexts: []string{"dev", "staging", "prod", "prod-eu", "prod-us"},
			include:  []string{},
			exclude:  []string{"prod"},
			want:     []string{"dev", "staging"},
		},
		{
			name:     "Include and exclude filters",
			contexts: []string{"dev", "staging", "prod", "prod-eu", "prod-us"},
			include:  []string{"prod"},
			exclude:  []string{"us"},
			want:     []string{"prod", "prod-eu"},
		},
		{
			name:     "Regex filter",
			contexts: []string{"dev", "staging", "prod", "prod-eu", "prod-us"},
			include:  []string{"/^prod-.*$/"},
			exclude:  []string{},
			want:     []string{"prod-eu", "prod-us"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterContexts(tt.contexts, tt.include, tt.exclude)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FilterContexts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		name    string
		s       string
		pattern string
		want    bool
	}{
		{
			name:    "Simple substring match - true",
			s:       "production",
			pattern: "prod",
			want:    true,
		},
		{
			name:    "Simple substring match - false",
			s:       "development",
			pattern: "prod",
			want:    false,
		},
		{
			name:    "Regex pattern match - true",
			s:       "prod-123",
			pattern: "/^prod-\\d+$/",
			want:    true,
		},
		{
			name:    "Regex pattern match - false",
			s:       "prod-abc",
			pattern: "/^prod-\\d+$/",
			want:    false,
		},
		{
			name:    "Invalid regex pattern",
			s:       "prod-123",
			pattern: "/^prod-[/", // Invalid regex
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchPattern(tt.s, tt.pattern); got != tt.want {
				t.Errorf("matchPattern() = %v, want %v", got, tt.want)
			}
		})
	}
}
