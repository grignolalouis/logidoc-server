package index

import (
	"encoding/json"
	"fmt"
	"strings"
)

func float64Ptr(f float64) *float64 { return &f }
func intPtr(i int) *int             { return &i }

// parseJSONArray finds and parses a JSON array from raw text that may contain
// surrounding non-JSON content. Handles truncated or malformed arrays commonly
// produced by LLMs by closing unbalanced brackets and trimming from the right.
func parseJSONArray[T any](raw string) ([]T, error) {
	raw = strings.TrimSpace(raw)

	start := strings.Index(raw, "[")
	if start < 0 {
		return nil, fmt.Errorf("no JSON array found")
	}
	sub := strings.TrimRight(raw[start:], " \t\n\r")

	var result []T
	if err := json.Unmarshal([]byte(sub), &result); err == nil && len(result) > 0 {
		return result, nil
	}

	if repaired := repairJSON(sub); repaired != sub {
		if err := json.Unmarshal([]byte(repaired), &result); err == nil && len(result) > 0 {
			return result, nil
		}
	}

	for trim := len(sub); trim > 0; trim-- {
		ch := sub[trim-1]
		if ch != ']' && ch != '}' {
			continue
		}
		var attempt []T
		if err := json.Unmarshal([]byte(sub[:trim]), &attempt); err == nil && len(attempt) > 0 {
			return attempt, nil
		}
		if repaired := repairJSON(sub[:trim]); repaired != sub[:trim] {
			if err := json.Unmarshal([]byte(repaired), &attempt); err == nil && len(attempt) > 0 {
				return attempt, nil
			}
		}
	}

	return nil, fmt.Errorf("invalid JSON array (len=%d)", len(sub))
}

// repairJSON closes any unclosed brackets and braces, ignoring those inside
// quoted strings.
func repairJSON(s string) string {
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

	for i := len(stack) - 1; i >= 0; i-- {
		s += string(stack[i])
	}
	return s
}
