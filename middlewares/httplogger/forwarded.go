package httplogger

import (
	"strings"
)

// ForwardedElement represents a single element in a Forwarded header.
type ForwardedElement struct {
	By    string
	For   string
	Host  string
	Proto string
}

// ParseForwarded parses the Forwarded header as defined in RFC 7239.
// It returns a slice of ForwardedElement, one for each element in the header.
func ParseForwarded(header string) []ForwardedElement {
	var elements []ForwardedElement

	if header == "" {
		return elements
	}

	// Split by comma to get individual elements
	parts := strings.Split(header, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		elem := ForwardedElement{}

		// Split by semicolon to get parameters
		params := strings.Split(part, ";")
		for _, param := range params {
			param = strings.TrimSpace(param)
			if param == "" {
				continue
			}

			key, value, found := strings.Cut(param, "=")
			if !found {
				continue
			}

			key = strings.ToLower(strings.TrimSpace(key))
			value = strings.TrimSpace(value)

			// Handle quoted strings
			if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
				value = value[1 : len(value)-1]
			}

			switch key {
			case "by":
				elem.By = value
			case "for":
				elem.For = value
			case "host":
				elem.Host = value
			case "proto":
				elem.Proto = value
			}
		}
		elements = append(elements, elem)
	}

	return elements
}
