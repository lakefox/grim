package parser

import (
	"grim/selector"
	"regexp"
	"strings"
)

type StyleMap struct {
	Selector    [][]string
	Styles      *map[string]string
	SheetNumber int
}

func ProcessStyles(selectString string) map[string]*StyleMap {
	sm := StyleMap{}
	styleMapMap := map[string]*StyleMap{}

	parts := strings.Split(selectString, ">")
	sm.Selector = make([][]string, len(parts))

	for i, v := range parts {
		part := selector.SplitSelector(strings.TrimSpace(v))
		sm.Selector[i] = part

		for _, b := range part {
			styleMapMap[b] = &sm
		}
	}
	return styleMapMap
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
		for _, selector := range selectors {
			styles := parseStyles(styleBlock)
			selectorMap[selector] = &styles
			smm := ProcessStyles(selector)
			for k := range smm {
				smm[k].Styles = &styles
				if styleMaps[k] == nil {
					styleMaps[k] = []*StyleMap{}
				}
				styleMaps[k] = append(styleMaps[k], smm[k])
			}
		}
	}

	return selectorMap, styleMaps
}

func parseSelectors(selectorBlock string) []string {
	// Split by comma and trim each selector
	selectors := strings.Split(selectorBlock, ",")
	for i, selector := range selectors {
		selectors[i] = strings.TrimSpace(selector)
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
