package parser

import (
	"grim/element"
	"regexp"
	"strings"
)

type StyleMap struct {
	Selector string
	Styles   *map[string]string
	Sheet int
}

func ParseCSS(css string, sheet int) map[string][]*StyleMap {
	// Remove comments
	css = removeComments(css)

	// Parse regular selectors and styles
	selectorRegex := regexp.MustCompile(`([^{]+){([^}]+)}`)
	matches := selectorRegex.FindAllStringSubmatch(css, -1)
	styleMaps := map[string][]*StyleMap{}
	for _, match := range matches {
		selectorBlock := strings.TrimSpace(match[1])
		styleBlock := match[2]

		selectors := parseSelectors(selectorBlock)
		for _, s := range selectors {
			styles := parseStyles(styleBlock)

			sel := element.ExtractBaseElements(s)
			smm := map[string]*StyleMap{}
			for _, is := range sel {
				for _, v := range is {
					smm[v] = &StyleMap{}
					smm[v].Styles = &styles
					smm[v].Selector = s
					if styleMaps[v] == nil {
						styleMaps[v] = []*StyleMap{}
					}
					styleMaps[v] = append(styleMaps[v], smm[v])
				}

			}
		}
	}

	return styleMaps
}

func parseSelectors(selectorBlock string) []string {
	var selectors []string
	var current strings.Builder
	var parenthesesDepth int

	for _, char := range selectorBlock {
		switch char {
		case '(':
			// Enter parentheses
			parenthesesDepth++
			current.WriteRune(char)
		case ')':
			// Exit parentheses
			parenthesesDepth--
			current.WriteRune(char)
		case ',':
			// Split only if not inside parentheses
			if parenthesesDepth == 0 {
				selectors = append(selectors, strings.TrimSpace(current.String()))
				current.Reset()
			} else {
				current.WriteRune(char)
			}
		default:
			current.WriteRune(char)
		}
	}

	// Add the last selector (if any)
	if current.Len() > 0 {
		selectors = append(selectors, strings.TrimSpace(current.String()))
	}

	return selectors
}

func parseStyles(styleBlock string) map[string]string {
	styleRegex := regexp.MustCompile(`([a-zA-Z-]+)\s*:\s*([^;]+);`)
	matches := styleRegex.FindAllStringSubmatch(styleBlock, -1)

	styleMap := make(map[string]string)
	for _, match := range matches {
		propName := strings.TrimSpace(match[1])
		propValue := strings.TrimSpace(match[2])
		styleMap[propName] = propValue
	}

	return styleMap
}

func ParseStyleAttribute(styleValue string) map[string]string {
	styleMap := make(map[string]string)

	start := 0
	for i := 0; i < len(styleValue); i++ {
		if styleValue[i] == ';' {
			part := styleValue[start:i]
			if len(part) > 0 {
				key, value := parseKeyValue(part)
				if key != "" && value != "" {
					styleMap[key] = value
				}
			}
			start = i + 1
		}
	}

	// Handle the last part if there's no trailing semicolon
	if start < len(styleValue) {
		part := styleValue[start:]
		key, value := parseKeyValue(part)
		if key != "" && value != "" {
			styleMap[key] = value
		}
	}

	return styleMap
}

func parseKeyValue(style string) (string, string) {
	for i := 0; i < len(style); i++ {
		if style[i] == ':' {
			key := strings.TrimSpace(style[:i])
			value := strings.TrimSpace(style[i+1:])
			return key, value
		}
	}
	return "", ""
}

func removeComments(css string) string {
	commentRegex := regexp.MustCompile(`(?s)/\*.*?\*/`)
	return commentRegex.ReplaceAllString(css, "")
}

