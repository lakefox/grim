package element

import (
	"strconv"
	"strings"
)

func Expander(styles map[string]string) map[string]string {
	parsed := parseBackground(styles["background"])
	delete(styles, "background")
	// Print result
	for key, value := range parsed {
		if value != "" {
			styles[key] = value
		}
	}

	if styles["margin"] != "" {
		left, right, top, bottom := convertMarginToIndividualProperties(styles["margin"])

		if styles["margin-left"] == "" {
			styles["margin-left"] = left
		}
		if styles["margin-right"] == "" {
			styles["margin-right"] = right
		}
		if styles["margin-top"] == "" {
			styles["margin-top"] = top
		}
		if styles["margin-bottom"] == "" {
			styles["margin-bottom"] = bottom
		}
	}
	delete(styles, "margin")

	if styles["padding"] != "" {
		left, right, top, bottom := convertMarginToIndividualProperties(styles["padding"])

		if styles["padding-left"] == "" {
			styles["padding-left"] = left
		}
		if styles["padding-right"] == "" {
			styles["padding-right"] = right
		}
		if styles["padding-top"] == "" {
			styles["padding-top"] = top
		}
		if styles["padding-bottom"] == "" {
			styles["padding-bottom"] = bottom
		}
	}
	delete(styles, "padding")

	flex := parseFlex(styles["flex"])

	delete(styles, "flex")

	styles["flex-basis"] = flex.FlexBasis
	styles["flex-grow"] = flex.FlexGrow
	styles["flex-shrink"] = flex.FlexShrink
	return styles
}

func convertMarginToIndividualProperties(margin string) (string, string, string, string) {
	parts := strings.Fields(margin)
	switch len(parts) {
	case 1:
		return parts[0], parts[0], parts[0], parts[0]
	case 2:
		return parts[0], parts[1], parts[0], parts[1]
	case 3:
		return parts[0], parts[1], parts[2], parts[1]
	case 4:
		return parts[0], parts[1], parts[2], parts[3]
	}
	return "0px", "0px", "0px", "0px"
}

// ParseBackground takes a CSS background shorthand and returns a map of its component parts.
func parseBackground(background string) map[string]string {
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

type FlexProperties struct {
	FlexGrow   string
	FlexShrink string
	FlexBasis  string
}

func parseFlex(flex string) FlexProperties {
	parts := strings.Fields(flex)
	prop := FlexProperties{
		FlexGrow:   "1",  // default value
		FlexShrink: "1",  // default value
		FlexBasis:  "0%", // default value
	}

	switch len(parts) {
	case 1:
		if strings.HasSuffix(parts[0], "%") || strings.HasSuffix(parts[0], "px") || strings.HasSuffix(parts[0], "em") {
			prop.FlexBasis = parts[0]
		} else if _, err := strconv.ParseFloat(parts[0], 64); err == nil {
			prop.FlexGrow = parts[0]
			prop.FlexShrink = "1"
			prop.FlexBasis = "0%"
		} else {
			return prop
		}
	case 2:
		prop.FlexGrow = parts[0]
		prop.FlexShrink = parts[1]
		prop.FlexBasis = "0%"
	case 3:
		prop.FlexGrow = parts[0]
		prop.FlexShrink = parts[1]
		prop.FlexBasis = parts[2]
	default:
		return prop
	}

	return prop
}
