package parser

import (
	"grim/element"
	"regexp"
	"strings"
)

type StyleMap struct {
	Selector string
	Styles   *map[string]string
}

func ParseCSS(css string) (map[string]*map[string]string, map[string][]*StyleMap) {
	selectorMap := make(map[string]*map[string]string)

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
			selectorMap[s] = &styles
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

	return selectorMap, styleMaps
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

// !TEMP: THIS IS TEMPERARY TO BY QS AND QSA WORK
func SplitSelector(s string) []string {
	var result []string
	var current strings.Builder
	var prePseudo string
	var pseudo string

	// Check if there's a `::` and split the string
	if idx := strings.Index(s, "::"); idx != -1 {
		prePseudo = s[:idx]
		pseudo = s[idx:] // Keep everything after `::` together
	} else {
		prePseudo = s
	}

	// Track whether we are inside parentheses
	var parenthesesDepth int

	for _, char := range prePseudo {
		switch char {
		case '(':
			// Enter parentheses
			parenthesesDepth++
			current.WriteRune(char)
		case ')':
			// Exit parentheses
			parenthesesDepth--
			current.WriteRune(char)
		case '.', '#', '[', ']', ':':
			if parenthesesDepth == 0 {
				// Only split outside parentheses
				if current.Len() > 0 {
					if char == ']' {
						current.WriteRune(char)
					}
					result = append(result, current.String())
					current.Reset()
				}
				if char != ']' {
					current.WriteRune(char)
				}
			} else {
				current.WriteRune(char)
			}
		default:
			current.WriteRune(char)
		}
	}

	if current.Len() > 0 {
		result = append(result, current.String())
	}

	// Add the pseudo-element (if any) as a single item
	if pseudo != "" {
		result = append(result, pseudo)
	}

	return result
}

func Contains(selector []string, node []string) bool {
	selectorSet := make(map[string]struct{}, len(node))
	for _, s := range node {
		selectorSet[strings.TrimSpace(s)] = struct{}{}
	}

	for _, s := range selector {
		if _, exists := selectorSet[strings.TrimSpace(s)]; !exists {
			return false
		}
	}
	return true
}
