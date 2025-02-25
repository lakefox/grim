package element

import (
	"strconv"
	"strings"
)

// !TODO: Make var() and :root (root is implide)

// !MAN,ELEMENT: QuerySelector is a Node method to select an element using a CSS selector
// + [!MAN]Note: the CSS query is performed relative to the Node calling it
// + [!MAN]Usage: element.Node.QuerySelector("css,query") -> element.Node
// + [!DEVMAN]Note: See TestSelector for query information
func (n *Node) QuerySelector(selectString string) *Node {
	m, _ := TestSelector(n, selectString)
	if m {
		return n
	}

	for i := range n.Children {
		el := n.Children[i]
		cr := el.QuerySelector(selectString)
		if cr.Properties.Id != "" {
			return cr
		}
	}

	return &Node{}
}

// !MAN,ELEMENT: QuerySelectorAll is a Node method to select all matching elements using a CSS selector
// + [!MAN]Note: the CSS query is performed relative to the Node calling it
// + [!MAN]Usage: element.Node.QuerySelectorAll("css,query") -> []element.Node
// + [!DEVMAN]Note: See TestSelector for query information
func (n *Node) QuerySelectorAll(selectString string) *[]*Node {
	results := []*Node{}

	m, _ := TestSelector(n, selectString)
	if m {
		results = append(results, n)
	}

	for i := range n.Children {
		el := n.Children[i]
		cr := el.QuerySelectorAll(selectString)
		if len(*cr) > 0 {
			results = append(results, *cr...)
		}
	}
	return &results
}

type SelectorParts struct {
	TagName   string
	Id        string
	ClassList []string
	Attribute map[string]string
}

func splitSelector(selector string, key rune) []string {
	var parts []string
	var current strings.Builder
	nesting := 0

	for _, r := range selector {
		switch r {
		case '(':
			nesting++
			current.WriteRune(r)
		case ')':
			if nesting > 0 {
				nesting--
			}
			current.WriteRune(r)
		case '[':
			nesting++
			current.WriteRune(r)
		case ']':
			if nesting > 0 {
				nesting--
			}
			current.WriteRune(r)
		case '{':
			nesting++
			current.WriteRune(r)
		case '}':
			if nesting > 0 {
				nesting--
			}
			current.WriteRune(r)
		case key:
			if nesting == 0 {
				// End of the current selector part
				parts = append(parts, strings.TrimSpace(current.String()))
				current.Reset()
			} else {
				// Inside nested context, add comma to the current part
				current.WriteRune(r)
			}
		default:
			current.WriteRune(r)
		}
	}

	// Add the last part if non-empty
	if current.Len() > 0 {
		parts = append(parts, strings.TrimSpace(current.String()))
	}

	return parts
}

// nthChildMatch checks if the given index matches the nth-child pattern.
func NthChildMatch(pattern string, index int) bool {
	pattern = strings.ReplaceAll(pattern, " ", "")
	// Handle special cases for "odd" and "even"
	lowerPattern := strings.ToLower(strings.TrimSpace(pattern))
	if lowerPattern == "odd" {
		return index%2 == 1
	}
	if lowerPattern == "even" {
		return index%2 == 0
	}

	// Coefficients for "an+b"
	a, b := 0, 0
	nIndex := strings.Index(lowerPattern, "n")

	// Parse pattern with 'n'
	if nIndex != -1 {
		// Parse coefficient of "n" (before 'n')
		if nIndex == 0 || lowerPattern[0] == '+' {
			a = 1
		} else if lowerPattern[0] == '-' && nIndex == 1 {
			a = -1
		} else {
			var err error
			a, err = strconv.Atoi(lowerPattern[:nIndex])
			if err != nil {
				return false
			}
		}

		// Parse constant term (after 'n')
		if nIndex+1 < len(lowerPattern) {
			b, _ = strconv.Atoi(lowerPattern[nIndex+1:])
		}
	} else {
		// Handle single integer patterns like "3"
		var err error
		b, err = strconv.Atoi(lowerPattern)
		if err != nil {
			return false
		}
	}

	// Check if index matches the formula a*n + b
	if a == 0 {
		return index == b
	}
	return (index-b)%a == 0 && (index-b)/a >= 0
}

// !ISSUE: :where doesn't work and h1+h1 is propagating to all elements on superselector
// + :where selects properly but the margin doesn't...
// + h1+h1 selects all but should only select the last two...

func TestSelector(n *Node, selector string) (bool, bool) {
	selectors := splitSelector(selector, ',')

	if selector[0] == ':' && len(selectors) == 1 {
		if selector == ":" {
			return true, true
		}
		if len(selector) >= 5 && selector[0:5] == ":has(" {
			m := false
			for _, v := range n.Children {
				m, _ = TestSelector(v, selector[5:len(selector)-1])
				if m {
					break
				}
			}
			return m, false
		} else if len(selector) >= 7 && selector[0:7] == ":where(" {
			m, _ := TestSelector(n, selector[7:len(selector)-1])
			return m, false
		} else if len(selector) >= 4 && selector[0:4] == ":is(" {
			m, _ := TestSelector(n, selector[4:len(selector)-1])
			return m, false
		} else if len(selector) >= 5 && selector[0:5] == ":not(" {
			m, _ := TestSelector(n, selector[5:len(selector)-1])
			return !m, false
		} else if len(selector) >= 11 && selector[0:11] == ":nth-child(" {
			index := 0
			if n.Parent != nil {
				for _, v := range n.Parent.Children {
					index++
					if v.Properties.Id == n.Properties.Id {
						break
					}
				}
			}
			m := NthChildMatch(selector[11:len(selector)-1], index)
			return m, false
		} else if selector == ":required" {
			return n.Required, false
		} else if selector == ":enabled" {
			return !n.Disabled, false
		} else if selector == ":disabled" {
			return n.Disabled, false
		} else if selector == ":checked" {
			return n.Checked, false
		} else if selector == ":focus" {
			return n.Focused, false
		} else if selector == ":hover" {
			return n.Hovered, false
		} else if selector == ":before" {
			return true, true
		} else if selector == ":after" {
			return true, true
		} else {
			return false, false
		}
	}

	has, isPsuedo := false, false
	for _, s := range selectors {
		directChildren := splitSelector(s, '>')
		if len(directChildren) > 1 {
			currentElement := n
			match := true
			for i := len(directChildren) - 1; i >= 0; i-- {
				m, _ := TestSelector(currentElement, directChildren[i])
				if !m {
					match = false
				}
				currentElement = currentElement.Parent
			}
			has = match
			break
		}
		for _, dc := range directChildren {
			adjacentSiblings := splitSelector(dc, '+')
			if len(adjacentSiblings) > 1 {
				for i, v := range n.Parent.Children {
					if v.Properties.Id == n.Properties.Id {
						// Make sure the current element matches the last selector
						m, _ := TestSelector(n, adjacentSiblings[len(adjacentSiblings)-1])
						if i >= len(adjacentSiblings)-1 && m {
							match := true
							// Skip the first selector, its been matched to the current node
							adjIndex := len(adjacentSiblings) - 2
							for j := i - 1; j >= i-len(adjacentSiblings)+1; j-- {
								sm, _ := TestSelector(n.Parent.Children[j], adjacentSiblings[adjIndex])
								if !sm {
									match = false
									break
								}
								adjIndex--
							}
							has = match
						}

						break
					}
				}

				break
			}
			for _, as := range adjacentSiblings {
				generalSiblings := splitSelector(as, '~')
				if len(generalSiblings) > 1 {
					match := false
					for _, sel := range generalSiblings {
						for i := 0; i < len(n.Parent.Children); i++ {
							v := n.Parent.Children[i]
							m, _ := TestSelector(v, sel)
							if m {
								if match {
									has, isPsuedo = true, false
									break
								}
								match = true
								break
							}
						}
						if !match {
							break
						}
					}
					has = match
					break
				}
				for _, gs := range generalSiblings {
					descendants := splitSelector(gs, ' ')
					if len(descendants) > 1 {
						// !DEVMAN,ELEMENT,TESTSELECTOR: Descendants selector checks for any parent to have
						// + the next tag, but doesn't check the main tag if it did and the main didn't
						// + have the selector then it would move up. So instead the main tag is skipped
						// + until the next check.
						currentElement := n.Parent
						match := false
						for i := len(descendants) - 1; i > 0; i-- {
							m := false
							for currentElement.Parent != nil && !m {
								m, _ = TestSelector(currentElement, descendants[i])
								if m {
									match = m
									break
								}
								currentElement = currentElement.Parent
							}

							if !m {
								match = false
								break
							}
						}
						if !match {
							break
						}
						has = match
						// A break is not inserted here because of the main element check
					}
					for _, d := range descendants {
						computeAble := splitSelector(d, ':')
						if len(computeAble) == 0 {
							continue
						}
						baseNode := computeAble[0]
						if len(computeAble) > 1 {
							match := true
							for _, v := range computeAble[1:] {
								m, p := TestSelector(n, ":"+v)
								isPsuedo = p
								if !m {
									match = false
								} else if len(v) > 10 && v[0:10] == "nth-child(" {
									has = CompareSelector(baseNode, n)
									break
								}
							}
							// !ISSUE: Why is this here??? has = is never reached above
							has = match
						}
						if has || !(len(computeAble) > 1) {
							m := CompareSelector(baseNode, n)
							has = m
						}
					}
				}
			}
		}
		if s == ":first-child" {
			has = n.Parent.Children[0].Properties.Id == n.Properties.Id
		} else if s == ":last-child" {
			has = n.Parent.Children[len(n.Parent.Children)-1].Properties.Id == n.Properties.Id
		}
		if has {
			return true, isPsuedo
		}
	}

	return has, isPsuedo
}

func ParseSelector(selector string) SelectorParts {
	parts := SelectorParts{
		Attribute: make(map[string]string),
	}

	// Parse character by character
	i := 0
	length := len(selector)

	// Parse tag name (if present)
	start := i
	for i < length && isValidTagChar(selector[i], i == start) {
		i++
	}
	if start < i {
		parts.TagName = selector[start:i]
	}

	// Parse the rest of the selector
	for i < length {
		switch selector[i] {
		case '#': // ID
			i++ // Skip #
			start = i
			for i < length && isValidIdClassChar(selector[i]) {
				i++
			}
			if start < i {
				parts.Id = selector[start:i]
			}

		case '.': // Class
			i++ // Skip .
			start = i
			for i < length && isValidIdClassChar(selector[i]) {
				i++
			}
			if start < i {
				parts.ClassList = append(parts.ClassList, selector[start:i])
			}

		case '[': // Attribute
			i++ // Skip [
			start = i
			var attrName string
			var attrValue string

			// Parse attribute name
			for i < length && selector[i] != '=' && selector[i] != ']' {
				i++
			}
			attrName = strings.TrimSpace(selector[start:i])

			// Parse attribute value if present
			if i < length && selector[i] == '=' {
				i++ // Skip =
				// Skip quotes if present
				if i < length && (selector[i] == '"' || selector[i] == '\'') {
					quote := selector[i]
					i++ // Skip opening quote
					start = i
					for i < length && selector[i] != quote {
						i++
					}
					attrValue = selector[start:i]
					if i < length {
						i++ // Skip closing quote
					}
				} else {
					start = i
					for i < length && selector[i] != ']' {
						i++
					}
					attrValue = strings.TrimSpace(selector[start:i])
				}
			}

			// Skip to closing bracket
			for i < length && selector[i] != ']' {
				i++
			}
			if i < length {
				i++ // Skip ]
			}

			parts.Attribute[attrName] = attrValue

		default:
			i++ // Skip any other character
		}
	}

	return parts
}

// Helper functions to check character validity
func isValidTagChar(c byte, isFirst bool) bool {
	if isFirst {
		return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
	}
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-'
}

func isValidIdClassChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-'
}

func CompareSelector(selector string, n *Node) bool {
	baseParts := ParseSelector(selector)
	has := false
	if selector == "*" {
		return true
	}
	if baseParts.Id == n.Id || baseParts.Id == "" {
		if baseParts.TagName == n.TagName || baseParts.TagName == "" {
			match := true
			for _, v := range baseParts.ClassList {
				bpc := false
				for _, c := range n.ClassList.Classes() {
					if v == c {
						bpc = true
					}
				}

				if !bpc {
					match = false
				}
			}

			for k, v := range baseParts.Attribute {
				if n.attribute[k] != v {
					match = false
				}
			}

			if match {
				has = true
			}
		}
	}
	return has
}

func ExtractBaseElements(selector string) [][]string {
	var baseElements []string

	selectors := splitSelector(selector, ',')
	for _, s := range selectors {
		directChild := splitSelector(s, '>')
		adjacentSibling := splitSelector(directChild[len(directChild)-1], '+')
		generalSiblings := splitSelector(adjacentSibling[len(adjacentSibling)-1], '~')
		descendants := splitSelector(generalSiblings[len(generalSiblings)-1], ' ')
		for _, d := range descendants {
			computeAble := splitSelector(d, ':')

			// Add valid base elements to the list
			if len(computeAble) > 0 && computeAble[0] != "" {
				baseElements = append(baseElements, computeAble[0])
			}
		}

	}

	baseParts := [][]string{}

	computeAble := splitSelector(selector, ':')

	for _, v := range computeAble {
		if strings.Contains(v, "(") {
			be := ExtractBaseElements(v[strings.Index(v, "(")+1 : len(v)-2])
			for _, b := range be {
				baseElements = append(baseElements, b...)
			}
		}
	}

	for _, v := range baseElements {
		ps := ParseSelector(v)
		selectors := []string{}
		selectors = append(selectors, ps.TagName)
		if ps.Id != "" {
			selectors = append(selectors, "#"+ps.Id)
		}

		for _, c := range ps.ClassList {
			selectors = append(selectors, "."+c)
		}

		for k, v := range ps.Attribute {
			selectors = append(selectors, `[`+k+`="`+v+`"]`)
		}

		baseParts = append(baseParts, selectors)
	}
	baseParts = append(baseParts, []string{"*"})
	return baseParts
}

func ShouldTestSelector(n *Node, selector string) bool {
	baseElements := ExtractBaseElements(selector)
	for _, base := range baseElements {
		if CompareSelector(strings.Join(base, ""), n) {
			return true
		}
	}
	return false
}

func GenBaseElements(n *Node) []string {
	selectors := []string{}
	selectors = append(selectors, n.TagName)
	if n.Id != "" {
		selectors = append(selectors, "#"+n.Id)
	}

	for _, c := range n.ClassList.Classes() {
		selectors = append(selectors, "."+c)
	}

	for k, v := range n.attribute {
		selectors = append(selectors, `[`+k+`="`+v+`"]`)
	}

	return selectors
}
