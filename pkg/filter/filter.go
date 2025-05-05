package filter

import (
	"regexp"
	"strings"
)

// FilterContexts filters kubernetes contexts based on include and exclude patterns
func FilterContexts(contexts []string, include, exclude []string) []string {
	if len(include) == 0 && len(exclude) == 0 {
		return contexts
	}

	result := make([]string, 0)

	for _, ctx := range contexts {
		if matchesFilters(ctx, include, exclude) {
			result = append(result, ctx)
		}
	}

	return result
}

// matchesFilters checks if a context matches the include filters and doesn't match the exclude filters
func matchesFilters(ctx string, include, exclude []string) bool {
	// If include filters are specified, context must match at least one
	if len(include) > 0 {
		matched := false
		for _, pattern := range include {
			if matchPattern(ctx, pattern) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// If exclude filters are specified, context must not match any
	for _, pattern := range exclude {
		if matchPattern(ctx, pattern) {
			return false
		}
	}

	return true
}

// matchPattern checks if a string matches a pattern (supporting basic glob and regex patterns)
func matchPattern(s, pattern string) bool {
	// Check for regex pattern (starts and ends with /)
	if strings.HasPrefix(pattern, "/") && strings.HasSuffix(pattern, "/") {
		regexPattern := pattern[1 : len(pattern)-1]
		reg, err := regexp.Compile(regexPattern)
		if err == nil && reg.MatchString(s) {
			return true
		}
		return false
	}

	// Simple substring matching
	return strings.Contains(s, pattern)
}
