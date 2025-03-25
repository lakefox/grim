package element

import (
	"bytes"
	"strings"
)

type StyleMap struct {
	Selector     string
	Styles       *map[string]string
	Sheet        int
	PsuedoStyles map[string]map[string]map[string]string
}

func ParseCSS(css string) map[string][]*StyleMap {
	// Remove comments
	css = removeComments(css)

	// Split into rule blocks
	blocks := splitBlocks(css)

	styleMaps := map[string][]*StyleMap{}

	for _, block := range blocks {
		// Split selector from style declarations
		parts := strings.SplitN(block, "{", 2)
		if len(parts) != 2 {
			continue // Invalid block
		}

		selectorBlock := strings.TrimSpace(parts[0])
		styleBlock := strings.TrimSpace(strings.TrimSuffix(parts[1], "}"))

		// Parse selectors and styles
		selectors := Token('(', ')', ',', selectorBlock)
		styles := parseStylesSimple(styleBlock)
		styles = Expander(styles)
		// Add to style maps
		for _, s := range selectors {
			sel := ExtractBaseElements(s)
			for _, is := range sel {
				for _, v := range is {
					styleMap := &StyleMap{
						Selector: s,
						Styles:   &styles,
					}

					if styleMaps[v] == nil {
						styleMaps[v] = []*StyleMap{}
					}
					styleMaps[v] = append(styleMaps[v], styleMap)
				}
			}
		}
	}

	return styleMaps
}

// splitBlocks splits CSS into rule blocks without using regex
func splitBlocks(css string) []string {
	var blocks []string
	var currentBlock bytes.Buffer
	var braceDepth int

	for i := 0; i < len(css); i++ {
		ch := css[i]

		if ch == '{' {
			braceDepth++
		} else if ch == '}' {
			braceDepth--
			if braceDepth == 0 {
				currentBlock.WriteByte(ch)
				blocks = append(blocks, currentBlock.String())
				currentBlock.Reset()
				continue
			}
		}

		if braceDepth > 0 || !isWhitespace(ch) {
			currentBlock.WriteByte(ch)
		}
	}

	return blocks
}

// parseStylesSimple parses CSS style declarations without using regex
func parseStylesSimple(styleBlock string) map[string]string {
	styles := make(map[string]string)
	declarations := strings.Split(styleBlock, ";")

	for _, declaration := range declarations {
		declaration = strings.TrimSpace(declaration)
		if declaration == "" {
			continue
		}

		parts := strings.SplitN(declaration, ":", 2)
		if len(parts) != 2 {
			continue
		}

		property := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if property != "" && value != "" {
			styles[property] = value
		}
	}

	return styles
}

// removeComments removes CSS comments without using regex
func removeComments(css string) string {
	var result bytes.Buffer
	i := 0

	for i < len(css) {
		if i+1 < len(css) && css[i] == '/' && css[i+1] == '*' {
			// Skip until end of comment
			i += 2
			for i+1 < len(css) && !(css[i] == '*' && css[i+1] == '/') {
				i++
			}
			i += 2 // Skip */
		} else {
			result.WriteByte(css[i])
			i++
		}
	}

	return result.String()
}

func isWhitespace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
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
	styleMap = Expander(styleMap)
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

// splitTokens splits a string into tokens, preserving content within parentheses
func Token(start, end, split rune, s string) []string {
	var tokens []string
	var currentToken strings.Builder
	parenDepth := 0
	inSpace := true // Track if we're in whitespace

	for _, char := range s {
		if char == start {
			parenDepth++
			currentToken.WriteRune(char)
			inSpace = false
		} else if char == end {
			parenDepth--
			currentToken.WriteRune(char)
			inSpace = false
		} else if char == split && parenDepth == 0 {
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
