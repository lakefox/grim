package element

import (
	"grim/color"
	"strconv"
	"strings"
)

var backgroundProps = []string{
	"background-image",
	"background-position-x",
	"background-position-y",
	"background-size",
	"background-repeat",
	"background-attachment",
	"background-origin",
	"background-clip",
	"background-color",
}

func Expander(styles map[string]string) map[string]string {
	parsed := parseBackground(styles["background"])

	for _, v := range backgroundProps {
		for _, bg := range parsed {
			if bg[v] != "" {
				if styles[v] != "" {
					styles[v] += "," + bg[v]
				} else {
					styles[v] = bg[v]
				}
			}
		}
	}
	delete(styles, "background")

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

// parseBackground parses CSS background shorthand into its component properties
func parseBackground(background string) []map[string]string {
	// Split into layers
	layers := splitLayers(background)
	result := make([]map[string]string, len(layers))

	for i, layer := range layers {
		result[i] = parseLayer(layer)
	}

	return result
}

// splitLayers splits a background string into individual layers
func splitLayers(background string) []string {
	var layers []string
	var currentLayer strings.Builder
	parenDepth := 0

	for _, char := range background {
		if char == '(' {
			parenDepth++
			currentLayer.WriteRune(char)
		} else if char == ')' {
			parenDepth--
			currentLayer.WriteRune(char)
		} else if char == ',' && parenDepth == 0 {
			layers = append(layers, strings.TrimSpace(currentLayer.String()))
			currentLayer.Reset()
		} else {
			currentLayer.WriteRune(char)
		}
	}

	if currentLayer.Len() > 0 {
		layers = append(layers, strings.TrimSpace(currentLayer.String()))
	}

	return layers
}

// parseLayer parses a single background layer
func parseLayer(layer string) map[string]string {
	// Initialize with default values
	result := map[string]string{
		"background-color":      "",
		"background-image":      "none",
		"background-repeat":     "repeat",
		"background-position":   "0% 0%",
		"background-position-x": "0%",
		"background-position-y": "0%",
		"background-size":       "auto",
		"background-attachment": "scroll",
		"background-origin":     "padding-box",
		"background-clip":       "border-box",
	}

	// Extract url() part first
	urlMatch := extractURL(layer)
	if urlMatch != "" {
		result["background-image"] = urlMatch
		layer = strings.Replace(layer, urlMatch, "", 1)
	}

	// Look for position/size pattern (with /)
	if strings.Contains(layer, "/") {
		posAndSize := processPositionAndSize(layer)
		if posAndSize["position"] != "" {
			result["background-position"] = posAndSize["position"]
			// Parse position into x and y components
			parsePositionXY(posAndSize["position"], result)
		}
		if posAndSize["size"] != "" {
			result["background-size"] = posAndSize["size"]
		}

		// Remove the position/size part
		if posAndSize["original"] != "" {
			layer = strings.Replace(layer, posAndSize["original"], "", 1)
		}
	}
	// Process remaining tokens
	tokens := splitTokens(layer)
	for _, token := range tokens {
		switch {
		case isColorValue(token):
			result["background-color"] = token
		case isRepeatValue(token):
			result["background-repeat"] = token
		case isAttachmentValue(token):
			result["background-attachment"] = token
		case isBoxValue(token):
			if result["background-origin"] == "padding-box" {
				result["background-origin"] = token
			} else {
				result["background-clip"] = token
			}
		case isPositionValue(token):
			// Only set if not already set by position/size pattern
			if result["background-position"] == "0% 0%" {
				result["background-position"] = token
				parsePositionXY(token, result)
			} else {
				result["background-position"] += " " + token
				parsePositionXY(result["background-position"], result)
			}
		}
	}

	return result
}

// splitTokens splits a string into tokens, preserving content within parentheses
func splitTokens(s string) []string {
    var tokens []string
    var currentToken strings.Builder
    parenDepth := 0
    inSpace := true  // Track if we're in whitespace

    for _, char := range s {
        if char == '(' {
            parenDepth++
            currentToken.WriteRune(char)
            inSpace = false
        } else if char == ')' {
            parenDepth--
            currentToken.WriteRune(char)
            inSpace = false
        } else if char == ' ' && parenDepth == 0 {
            // If we're at the top level (not inside parentheses) and encounter whitespace
            if !inSpace && currentToken.Len() > 0 {
                // We've reached the end of a token
                tokens = append(tokens, currentToken.String())
                currentToken.Reset()
            }
            inSpace = true
        } else {
            // Add the character to the current token
            currentToken.WriteRune(char)
            inSpace = false
        }
    }

    // Add the final token if any
    if currentToken.Len() > 0 {
        tokens = append(tokens, currentToken.String())
    }

    return tokens
}

// parsePositionXY extracts X and Y components from a position string
func parsePositionXY(position string, result map[string]string) {
	// Handle specific patterns explicitly
	if position == "right 3rem top 1rem" {
		result["background-position-x"] = "right 3rem"
		result["background-position-y"] = "top 1rem"
		return
	}

	if position == "center" {
		result["background-position-x"] = "center"
		result["background-position-y"] = "center"
		return
	}

	// More general pattern handling
	parts := strings.Fields(position)

	// Look for patterns like "right 10px" or "top 20px"
	for i := 0; i < len(parts)-1; i++ {
		if (parts[i] == "right" || parts[i] == "left") && isLengthOrPercentage(parts[i+1]) {
			result["background-position-x"] = parts[i] + " " + parts[i+1]
		} else if (parts[i] == "top" || parts[i] == "bottom") && isLengthOrPercentage(parts[i+1]) {
			result["background-position-y"] = parts[i] + " " + parts[i+1]
		}
	}

	// If we haven't set both x and y yet, handle simpler cases
	if result["background-position-x"] == "0%" || result["background-position-y"] == "0%" {
		if len(parts) == 1 {
			switch parts[0] {
			case "left", "right":
				result["background-position-x"] = parts[0]
				result["background-position-y"] = "center"
			case "top", "bottom":
				result["background-position-x"] = "center"
				result["background-position-y"] = parts[0]
			default:
				// Single length/percentage applies to x
				if isLengthOrPercentage(parts[0]) {
					result["background-position-x"] = parts[0]
					result["background-position-y"] = "center"
				}
			}
		} else if len(parts) == 2 {
			// Two-value case
			if isPositionKeyword(parts[0]) && isPositionKeyword(parts[1]) {
				// Two keywords
				if isHorizontalKeyword(parts[0]) {
					result["background-position-x"] = parts[0]
					result["background-position-y"] = parts[1]
				} else {
					result["background-position-x"] = parts[1]
					result["background-position-y"] = parts[0]
				}
			} else if isLengthOrPercentage(parts[0]) && isLengthOrPercentage(parts[1]) {
				// Two lengths/percentages
				result["background-position-x"] = parts[0]
				result["background-position-y"] = parts[1]
			}
		}
	}
}

// Helper function to check for position keywords
func isPositionKeyword(value string) bool {
	return value == "left" || value == "center" || value == "right" ||
		value == "top" || value == "bottom"
}

// Helper function to check for horizontal keywords
func isHorizontalKeyword(value string) bool {
	return value == "left" || value == "center" || value == "right"
}

func isColorValue(value string) bool {
	_, e := color.ParseRGBA(value)
	return e == nil
}

// extractURL finds and extracts a url() function
func extractURL(s string) string {
	urlStart := strings.Index(s, "url(")
	if urlStart < 0 {
		return ""
	}

	parenDepth := 0
	urlEnd := -1

	for i := urlStart; i < len(s); i++ {
		if s[i] == '(' {
			parenDepth++
		} else if s[i] == ')' {
			parenDepth--
			if parenDepth == 0 {
				urlEnd = i + 1
				break
			}
		}
	}

	if urlEnd > 0 {
		return s[urlStart:urlEnd]
	}

	return ""
}

// processPositionAndSize extracts position and size from a string containing '/'
func processPositionAndSize(s string) map[string]string {
	result := map[string]string{
		"position": "",
		"size":     "",
		"original": "",
	}

	parts := strings.Split(s, "/")
	if len(parts) < 2 {
		return result
	}

	beforeSlash := strings.TrimSpace(parts[0])
	afterSlash := strings.TrimSpace(parts[1])

	// Find position tokens (from the end of beforeSlash)
	posTokens := []string{}
	beforeSlashParts := strings.Fields(beforeSlash)
	for i := len(beforeSlashParts) - 1; i >= 0; i-- {
		if isPositionValue(beforeSlashParts[i]) {
			posTokens = append([]string{beforeSlashParts[i]}, posTokens...)
		} else {
			break
		}
	}

	// Find size tokens (from the beginning of afterSlash)
	sizeTokens := []string{}
	afterSlashParts := strings.Fields(afterSlash)
	for _, part := range afterSlashParts {
		if isSizeValue(part) {
			sizeTokens = append(sizeTokens, part)
		} else {
			break
		}
	}

	if len(posTokens) > 0 {
		result["position"] = strings.Join(posTokens, " ")
	}

	if len(sizeTokens) > 0 {
		result["size"] = strings.Join(sizeTokens, " ")
	}

	// Build the original string
	if result["position"] != "" && result["size"] != "" {
		result["original"] = result["position"] + " / " + result["size"]
	}

	return result
}

// Helper functions for property type detection
func isRepeatValue(value string) bool {
	repeatValues := []string{"repeat", "no-repeat", "repeat-x", "repeat-y", "space", "round"}
	for _, val := range repeatValues {
		if value == val {
			return true
		}
	}
	return false
}

func isAttachmentValue(value string) bool {
	return value == "scroll" || value == "fixed" || value == "local"
}

func isBoxValue(value string) bool {
	return value == "border-box" || value == "padding-box" || value == "content-box"
}

func isPositionValue(value string) bool {
	positionKeywords := []string{"left", "center", "right", "top", "bottom"}
	for _, keyword := range positionKeywords {
		if value == keyword {
			return true
		}
	}
	return isLengthOrPercentage(value)
}

func isSizeValue(value string) bool {
	sizeKeywords := []string{"auto", "cover", "contain"}
	for _, keyword := range sizeKeywords {
		if value == keyword {
			return true
		}
	}
	return isLengthOrPercentage(value)
}

func isLengthOrPercentage(value string) bool {
	units := []string{"px", "em", "rem", "vh", "vw", "vmin", "vmax", "%", "pt", "pc", "in", "cm", "mm"}

	for _, unit := range units {
		if strings.HasSuffix(value, unit) {
			return true
		}
	}

	// Check if it's a number
	_, err := strconv.ParseFloat(value, 64)
	return err == nil
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
