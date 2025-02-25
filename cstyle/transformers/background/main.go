package background

import (
	"grim/cstyle"
	"grim/element"
	"strings"
)

func Init() cstyle.Transformer {
	return cstyle.Transformer{
		Selector: func(n *element.Node, c *cstyle.CSS) bool {
			return n.Style("background") != ""
		},
		Handler: func(n *element.Node, c *cstyle.CSS) *element.Node {
			parsed := ParseBackground(n.Style("background"))

			// Print result
			for key, value := range parsed {
				if value != "" {
					n.Style(key, value)
				}
			}

			return n
		},
	}
}

// ParseBackground takes a CSS background shorthand and returns a map of its component parts.
func ParseBackground(background string) map[string]string {
	parts := splitBackground(background)
	result := make(map[string]string)

	// Default component properties
	result["background-color"] = ""
	result["background-image"] = "none"
	result["background-repeat"] = "repeat"
	result["background-position"] = "0% 0%"
	result["background-size"] = "auto"
	result["background-attachment"] = "scroll"
	result["background-origin"] = "padding-box"
	result["background-clip"] = "border-box"

	for _, part := range parts {
		switch {
		// Handle background-image (assuming url format)
		case strings.HasPrefix(part, "url("):
			result["background-image"] = part

		// Handle background-repeat (no-repeat, repeat-x, repeat-y)
		case part == "no-repeat" || part == "repeat" || part == "repeat-x" || part == "repeat-y":
			result["background-repeat"] = part

		// Handle background-attachment (scroll or fixed)
		case part == "scroll" || part == "fixed":
			result["background-attachment"] = part

		// Handle background-position (percentage or predefined values)
		case strings.Contains(part, "%") || isPosition(part):
			result["background-position"] = part

		// Handle background-size (contain, cover, or specific size)
		case part == "contain" || part == "cover" || strings.Contains(part, "px") || strings.Contains(part, "%"):
			result["background-size"] = part

		// Handle background-origin (border-box, padding-box, content-box)
		case part == "border-box" || part == "padding-box" || part == "content-box":
			result["background-origin"] = part
			result["background-clip"] = part // background-clip defaults to the same as background-origin

		// Handle background-color (rgb, rgba, hsl, hsla)
		case isColorFunction(part):
			result["background-color"] = part

		// Handle background-color for basic colors or unknown values
		default:
			result["background-color"] = part
		}
	}

	return result
}

// splitBackground splits background properties while preserving functions like rgb(), rgba(), hsl(), etc.
func splitBackground(background string) []string {
	var result []string
	var current strings.Builder
	parenDepth := 0
	inWord := false

	for _, char := range background {
		// Track parentheses depth
		if char == '(' {
			parenDepth++
			current.WriteRune(char)
			inWord = true
			continue
		}
		if char == ')' {
			parenDepth--
			current.WriteRune(char)
			inWord = true
			continue
		}

		// Handle spaces - they're separators only when not inside parentheses
		if char == ' ' || char == '\t' {
			if parenDepth > 0 {
				// Inside parentheses, preserve the space
				current.WriteRune(char)
			} else if inWord {
				// End of a word, add to results
				result = append(result, current.String())
				current.Reset()
				inWord = false
			}
			continue
		}

		// Any other character
		current.WriteRune(char)
		inWord = true
	}

	// Add the last part if there is one
	if current.Len() > 0 {
		result = append(result, current.String())
	}

	return result
}

// Helper to check if a string is a valid CSS color function (e.g., rgb(), rgba(), hsl(), hsla())
func isColorFunction(value string) bool {
	// Check for rgb(), rgba(), hsl(), or hsla() functions
	return strings.HasPrefix(value, "rgb(") ||
		strings.HasPrefix(value, "rgba(") ||
		strings.HasPrefix(value, "hsl(") ||
		strings.HasPrefix(value, "hsla(")
}

// Helper to check if a string is a valid background position
func isPosition(value string) bool {
	positions := []string{"left", "right", "top", "bottom", "center"}
	for _, pos := range positions {
		if value == pos {
			return true
		}
	}
	return false
}
