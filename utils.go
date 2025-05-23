package grim

import (
	"fmt"
	ic "image/color"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

func GetXY(n Node, state map[string]State) (float32, float32) {
	s := state
	// self := s[n.Properties.Id]

	offsetX := float32(0)
	offsetY := float32(0)
	p := n.parent
	if p != nil {
		parent := s[p.Properties.Id]
		offsetX += parent.Border.Left.Width + parent.Padding.Left
		offsetY += parent.Border.Top.Width + parent.Padding.Top
	}

	return offsetX, offsetY
}

type BoxSizing struct {
	Width  float32
	Height float32
}

func FindBounds(n Node, style map[string]string, state *map[string]State) (BoxSizing, BoxSpacing, BoxSpacing) {
	s := *state
	self := s[n.Properties.Id]
	var parent State

	fs := self.EM

	p := n.parent
	var pwh BoxSizing
	if p != nil {
		parent = s[p.Properties.Id]
		pwh = BoxSizing{
			Width:  parent.Width,
			Height: parent.Height,
		}
	} else {
		pwh = BoxSizing{}
		if width, exists := style["width"]; exists {
			if f, err := strconv.ParseFloat(strings.TrimSuffix(width, "px"), 32); err == nil {
				pwh.Width = float32(f)
			}
		}
		if height, exists := style["height"]; exists {
			if f, err := strconv.ParseFloat(strings.TrimSuffix(height, "px"), 32); err == nil {
				pwh.Height = float32(f)
			}
		}
	}

	wStyle := style["width"]

	if wStyle == "" && style["display"] != "inline" {
		wStyle = "100%"
	}

	width := ConvertToPixels(wStyle, fs, pwh.Width)
	height := ConvertToPixels(style["height"], fs, pwh.Height)

	if minWidth, exists := style["min-width"]; exists {
		width = Max(width, ConvertToPixels(minWidth, fs, pwh.Width))
	}
	if maxWidth, exists := style["max-width"]; exists {
		width = Min(width, ConvertToPixels(maxWidth, fs, pwh.Width))
	}
	if minHeight, exists := style["min-height"]; exists {
		height = Max(height, ConvertToPixels(minHeight, fs, pwh.Height))
	}
	if maxHeight, exists := style["max-height"]; exists {
		height = Min(height, ConvertToPixels(maxHeight, fs, pwh.Height))
	}

	wh := BoxSizing{
		Width:  width,
		Height: height,
	}

	m := getMP(n, style, wh, state, "margin")
	padding := getMP(n, style, wh, state, "padding")

	if p != nil {
		wh.Width += padding.Left + padding.Right
		wh.Height += padding.Top + padding.Bottom
	}

	if wStyle == "100%" && style["position"] != "absolute" {
		wh.Width -= (m.Right + m.Left + self.Border.Left.Width + self.Border.Right.Width + parent.Padding.Left + parent.Padding.Right + padding.Left + padding.Right)
	}

	if style["height"] == "100%" {
		if style["position"] == "absolute" {
			wh.Height -= (m.Top + m.Bottom)
		} else {
			wh.Height -= (m.Top + m.Bottom + parent.Padding.Top + parent.Padding.Bottom)
		}
	}

	return wh, m, padding
}

func getMP(n Node, style map[string]string, wh BoxSizing, state *map[string]State, t string) BoxSpacing {
	s := *state
	self := s[n.Properties.Id]
	fs := self.EM
	m := BoxSpacing{}

	// Cache style properties
	leftKey, rightKey, topKey, bottomKey := t+"-left", t+"-right", t+"-top", t+"-bottom"

	leftStyle := style[leftKey]
	rightStyle := style[rightKey]
	topStyle := style[topKey]
	bottomStyle := style[bottomKey]

	// Convert left and right properties
	if leftStyle != "" || rightStyle != "" {
		m.Left = ConvertToPixels(leftStyle, fs, wh.Width)
		m.Right = ConvertToPixels(rightStyle, fs, wh.Width)
	}

	// Convert top and bottom properties
	if topStyle != "" || bottomStyle != "" {
		m.Top = ConvertToPixels(topStyle, fs, wh.Height)
		m.Bottom = ConvertToPixels(bottomStyle, fs, wh.Height)
	}

	p := n.parent

	if t == "margin" {
		siblingMargin := float32(0)
		firstChild := false
		// Margin Collapse
		// !ISSUE: Check margin collapse
		if p != nil {
			sibIndex := -1
			for i, v := range p.Children {
				if v.Properties.Id == n.Properties.Id {
					sibIndex = i - 1

					break
				}
			}
			if sibIndex > -1 {
				sib := s[p.Children[sibIndex].Properties.Id]
				siblingMargin = sib.Margin.Bottom
			}
		}

		// Handle top margin collapse
		for i, v := range p.Children {
			if v.Properties.Id == n.Properties.Id {
				if i == 0 {
					firstChild = true
				}
				break
			}
		}
		if firstChild {
			parent := s[p.Properties.Id]
			if parent.Margin.Top < m.Top {
				parent.Margin.Top = m.Top
				(*state)[p.Properties.Id] = parent
			}
			m.Top = 0
		} else {
			if m.Top != 0 {
				if m.Top < 0 {
					m.Top += siblingMargin
				} else {
					m.Top = Max(m.Top-siblingMargin, 0)
				}
			}
		}

		// Handle auto margins
		if leftStyle == "auto" && rightStyle == "auto" {
			parent := s[p.Properties.Id]
			pwh := BoxSizing{
				Width: parent.Width,
			}
			m.Left = Max((pwh.Width-wh.Width)/2, 0)
			m.Right = m.Left
		}
	}

	return m
}

var unitFactors = map[string]float64{
	"px":   1,
	"rem":  -1, // special handling
	"em":   -1, // special handling
	"pt":   1.33,
	"pc":   16.89,
	"%":    -1, // special handling
	"vw":   -1, // special handling
	"vh":   -1, // special handling
	"cm":   37.79527559,
	"in":   96,
	"auto": -1,
}

// ConvertToPixels converts a CSS measurement to pixels.
func ConvertToPixels(value string, em, max float32) float32 {
	// Quick check for predefined units
	switch value {
	case "thick":
		return 5
	case "medium":
		return 3
	case "thin":
		return 1
	}

	// Handle calculation expression
	if len(value) > 5 && value[:5] == "calc(" {
		return evaluateCalcExpression(value[5:len(value)-1], em, max)
	}

	for unit, factor := range unitFactors {
		if strings.HasSuffix(value, unit) {
			if unit == "em" && strings.HasSuffix(value, "rem") {
				continue
			}
			cutStr := strings.TrimSuffix(value, unit)
			numericValue, err := strconv.ParseFloat(cutStr, 64)
			if err != nil && value != "auto" {
				return 0
			}
			// Handle special units like "em", "%" etc.
			if factor == -1 {
				switch unit {
				case "em":
					return float32(numericValue) * em
				// !ISSUE: REM not properly impleamented
				case "rem":
					return float32(numericValue) * em
				case "%", "vw", "vh":
					return float32(numericValue) * (max / 100)
				case "pt":
					return (float32(numericValue) * 96) / 72
				case "auto":
					return max
				}
			}
			return float32(numericValue) * float32(factor)
		}
	}

	// Default return if no match
	return 0
}

// evaluateCalcExpression recursively evaluates 'calc()' expressions
func evaluateCalcExpression(expression string, em, max float32) float32 {
	terms := strings.FieldsFunc(expression, func(c rune) bool {
		return c == '+' || c == '-' || c == '*' || c == '/'
	})

	operators := strings.FieldsFunc(expression, func(c rune) bool {
		return c != '+' && c != '-' && c != '*' && c != '/'
	})

	var result float32

	for i, term := range terms {
		value := ConvertToPixels(strings.TrimSpace(term), em, max)

		if i > 0 {
			switch operators[i-1] {
			case "+":
				result += value
			case "-":
				result -= value
			case "*":
				result *= value
			case "/":
				if value != 0 {
					result /= value
				} else {
					return 0
				}
			}
		} else {
			result = value
		}
	}

	return result
}

func Max(a, b float32) float32 {
	if a > b {
		return a
	} else {
		return b
	}
}

func Min(a, b float32) float32 {
	if a < b {
		return a
	} else {
		return b
	}
}

func GetInnerText(n *html.Node) string {
	var result strings.Builder

	var getText func(*html.Node)
	getText = func(n *html.Node) {
		// Skip processing if the node is a head tag
		if n.Type == html.ElementNode && n.Data == "head" {
			return
		}

		// If it's a text node, append its content
		if n.Type == html.TextNode {
			result.WriteString(n.Data)
		}

		// Traverse child nodes recursively
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			getText(c)
		}
	}

	getText(n)

	return result.String()
}

func RGBAtoString(c ic.RGBA) string {
	return fmt.Sprintf("R%d%d%d%d", c.R, c.G, c.B, c.A)
}

func SplitByComma(input string) []string {
	var result []string
	var current strings.Builder

	// Track nesting level for each bracket type
	squareBrackets := 0 // []
	curlyBraces := 0    // {}
	parentheses := 0    // ()

	for _, char := range input {
		switch char {
		case '[':
			squareBrackets++
			current.WriteRune(char)
		case ']':
			squareBrackets--
			current.WriteRune(char)
		case '{':
			curlyBraces++
			current.WriteRune(char)
		case '}':
			curlyBraces--
			current.WriteRune(char)
		case '(':
			parentheses++
			current.WriteRune(char)
		case ')':
			parentheses--
			current.WriteRune(char)
		case ',':
			// Only split on comma if we're not inside any brackets
			if squareBrackets == 0 && curlyBraces == 0 && parentheses == 0 {
				result = append(result, strings.TrimSpace(current.String()))
				current.Reset()
			} else {
				current.WriteRune(char)
			}
		default:
			current.WriteRune(char)
		}
	}

	// Add the last segment if there's anything left
	if current.Len() > 0 {
		result = append(result, strings.TrimSpace(current.String()))
	}

	return result
}
