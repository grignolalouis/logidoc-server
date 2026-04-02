// Package jsonutil provides helpers for parsing and repairing LLM-generated JSON.
package jsonutil

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ParseArray finds and parses a JSON array from raw text that may contain
// surrounding non-JSON content. Handles truncated or malformed arrays
// commonly produced by LLMs.
func ParseArray[T any](raw string) ([]T, error) {
	raw = strings.TrimSpace(raw)

	start := strings.Index(raw, "[")
	if start < 0 {
		return nil, fmt.Errorf("no JSON array found")
	}
	sub := strings.TrimRight(raw[start:], " \t\n\r")

	// Direct parse
	var result []T
	if err := json.Unmarshal([]byte(sub), &result); err == nil && len(result) > 0 {
		return result, nil
	}

	// Try repairing unclosed brackets
	if repaired := Repair(sub); repaired != sub {
		if err := json.Unmarshal([]byte(repaired), &result); err == nil && len(result) > 0 {
			return result, nil
		}
	}

	// Try trimming trailing chars
	for trim := len(sub); trim > 0; trim-- {
		ch := sub[trim-1]
		if ch != ']' && ch != '}' {
			continue
		}
		var attempt []T
		if err := json.Unmarshal([]byte(sub[:trim]), &attempt); err == nil && len(attempt) > 0 {
			return attempt, nil
		}
		if repaired := Repair(sub[:trim]); repaired != sub[:trim] {
			if err := json.Unmarshal([]byte(repaired), &attempt); err == nil && len(attempt) > 0 {
				return attempt, nil
			}
		}
	}

	return nil, fmt.Errorf("invalid JSON array (len=%d)", len(sub))
}

// Repair closes any unclosed brackets and braces in a JSON string.
// Handles strings correctly (doesn't count brackets inside quoted strings).
func Repair(s string) string {
	var stack []byte
	inString, escape := false, false

	for _, c := range []byte(s) {
		if escape {
			escape = false
			continue
		}
		if c == '\\' && inString {
			escape = true
			continue
		}
		if c == '"' {
			inString = !inString
			continue
		}
		if inString {
			continue
		}
		switch c {
		case '{':
			stack = append(stack, '}')
		case '[':
			stack = append(stack, ']')
		case '}', ']':
			if len(stack) > 0 && stack[len(stack)-1] == c {
				stack = stack[:len(stack)-1]
			}
		}
	}

	// Close unclosed brackets in reverse
	for i := len(stack) - 1; i >= 0; i-- {
		s += string(stack[i])
	}
	return s
}

// Truncate returns the first n characters of s, appending "..." if truncated.
func Truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
