package selector

import (
	"fmt"
	"grim/element"
	"regexp"
	"strconv"
	"strings"
)

// !TODO: Create :not and other selectors
// + :nth-last-child()etc..

// !TODO: Make var() and :root (root is implide)

// !ISSUE: Psuedo selectors don't work

type SelectorParts struct {
	TagName    string
	ID         string
	Classes    []string
	Attributes map[string]string
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

func TestSelector(n *element.Node, selector string) bool {
	selectors := splitSelector(selector, ',')
	// fmt.Println(selectors)
	// fmt.Println("START: ", n.Properties.Id, selector)

	if selector[0] == ':' {

		if len(selector) >= 5 && selector[0:5] == ":has(" {
			m := false
			for _, v := range n.Children {
				m = TestSelector(v, selector[5:len(selector)-1])
				// fmt.Println("HAS", selector[5:len(selector)-1], m)
				if m {
					break
				}
			}
			return m
		} else if len(selector) >= 7 && selector[0:7] == ":where(" {
			m := TestSelector(n, selector[7:len(selector)-1])
			// fmt.Println("WHERE", selector[7:len(selector)-1], m)
			return m
		} else if len(selector) >= 4 && selector[0:4] == ":is(" {
			m := TestSelector(n, selector[4:len(selector)-1])
			// fmt.Println("IS: ", selector[4:len(selector)-1], m)
			return m
		} else if len(selector) >= 5 && selector[0:5] == ":not(" {
			m := !TestSelector(n, selector[5:len(selector)-1])
			// fmt.Println("NOT: ", selector[5:len(selector)-1], m)
			return m
		} else if len(selector) >= 11 && selector[0:11] == ":nth-child(" {
			index := 0
			for _, v := range n.Parent.Children {
				index++
				if v.Properties.Id == n.Properties.Id {
					break
				}
			}
			m := NthChildMatch(selector[11:len(selector)-1], index)
			// fmt.Println("NTH: ", selector[11:len(selector)-1], m)
			return m
		} else if selector == ":required" {
			return n.Required
		} else if selector == ":enabled" {
			return !n.Disabled
		} else if selector == ":disabled" {
			return n.Disabled
		} else if selector == ":checked" {
			return n.Checked
		} else if selector == ":focus" {
			return n.Focused
		} else if selector == ":hover" {
			return n.Hovered
		}
	}
	has := false
	for _, s := range selectors {
		directChildren := splitSelector(s, '>')
		if len(directChildren) > 1 {
			currentElement := n
			for i := len(directChildren) - 2; i >= 0; i-- {
				currentElement = currentElement.Parent
				if TestSelector(currentElement, directChildren[i]) {
					has = true
					break
				}
			}
		}
		for _, dc := range directChildren {
			adjacentSiblings := splitSelector(dc, '+')
			if len(adjacentSiblings) > 1 {
				// fmt.Println("ADJ: ", adjacentSiblings, len(adjacentSiblings), n.Properties.Id)
				match := false
				index := 0
				for _, sel := range adjacentSiblings {
					// fmt.Println("TESTING ADJ: ", sel, len(n.Parent.Children), index)
					for i := index; i < len(n.Parent.Children); i++ {
						v := n.Parent.Children[i]
						// fmt.Println("ON: ", v.Properties.Id)
						if TestSelector(v, sel) {
							if match {
								fmt.Println("SIBLINGS: ", i == index, i, index)
								return i == index
							}
							match = true
							index = i + 1
							if index > len(n.Parent.Children)-1 {
								has = false
								break
							}
							break
						}
					}
					if !match {
						break
					}
				}
				// fmt.Println("ADJ RES: ", match)
				has = match
				break
			}
			for _, as := range adjacentSiblings {
				generalSiblings := splitSelector(as, '~')
				if len(generalSiblings) > 1 {
					// fmt.Println("GEN SIB: ", generalSiblings, len(generalSiblings), n.Properties.Id)
					match := false
					for _, sel := range generalSiblings {
						// fmt.Println("TESTING GEN SIB: ", sel, len(n.Parent.Children))
						for i := 0; i < len(n.Parent.Children); i++ {
							v := n.Parent.Children[i]
							// fmt.Println("ON: ", v.Properties.Id)
							if TestSelector(v, sel) {
								if match {
									return true
								}
								match = true
								break
							}
						}
						if !match {
							break
						}
					}
					// fmt.Println("GEN SIB RES: ", match)
					has = match
					break
				}
				for _, gs := range generalSiblings {
					descendants := splitSelector(gs, ' ')
					for _, d := range descendants {
						computeAble := splitSelector(d, ':')
						baseNode := computeAble[0]
						// fmt.Println("BN", baseNode, computeAble, len(computeAble))

						if len(computeAble) > 1 {
							match := true
							for _, v := range computeAble[1:] {
								// fmt.Println("TESTING COMPUTABLE")
								m := TestSelector(n, ":"+v)
								if !m {
									match = false
								}
							}
							has = match
						}
						if has || !(len(computeAble) > 1) {
							m := CompareSelector(baseNode, n)
							// fmt.Println("NO COMPUTABLE: ", m, baseNode, n.Properties.Id)
							has = m
						}
					}
				}
			}
		}
		if has {
			return true
		}
	}
	return has
}

func ParseSelector(selector string) SelectorParts {
	// Initialize parts
	parts := SelectorParts{
		Attributes: make(map[string]string),
	}

	// Regular expressions for components
	tagRegex := `^[a-zA-Z][a-zA-Z0-9-]*`
	idRegex := `#([a-zA-Z0-9_-]+)`
	classRegex := `\.([a-zA-Z0-9_-]+)`
	attrRegex := `\[(.+?)\]`

	// Extract tag name
	tagMatch := regexp.MustCompile(tagRegex).FindString(selector)
	if tagMatch != "" {
		parts.TagName = tagMatch
		selector = strings.TrimPrefix(selector, tagMatch)
	}

	// Extract attributes
	attrMatches := regexp.MustCompile(attrRegex).FindAllStringSubmatch(selector, -1)
	for _, match := range attrMatches {
		attr := match[1]
		kv := strings.SplitN(attr, "=", 2)
		if len(kv) == 2 {
			key := kv[0]
			value := strings.Trim(kv[1], `"`)
			parts.Attributes[key] = value
		} else {
			// Handle attributes without values
			parts.Attributes[kv[0]] = ""
		}
		selector = strings.Replace(selector, match[0], "", 1)
	}

	// Extract ID
	idMatch := regexp.MustCompile(idRegex).FindStringSubmatch(selector)
	if len(idMatch) > 1 {
		parts.ID = idMatch[1]
		selector = strings.Replace(selector, idMatch[0], "", 1)
	}

	// Extract classes
	classMatches := regexp.MustCompile(classRegex).FindAllStringSubmatch(selector, -1)
	for _, match := range classMatches {
		parts.Classes = append(parts.Classes, match[1])
	}

	return parts
}

func CompareSelector(selector string, n *element.Node) bool {
	baseParts := ParseSelector(selector)
	has := false
	if baseParts.ID == n.Id || baseParts.ID == "" {
		if baseParts.TagName == n.TagName || baseParts.TagName == "" {
			match := true
			for _, v := range baseParts.Classes {
				bpc := false
				for _, c := range n.ClassList.Classes {
					if v == c {
						bpc = true
					}
				}

				if !bpc {
					match = false
				}
			}

			for k, v := range baseParts.Attributes {
				if n.Attribute[k] != v {
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
