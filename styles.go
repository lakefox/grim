package grim

import (
	"sort"
	"strconv"
	"strings"
)

type Styles struct {
	StyleMap     map[string][]*StyleMap
	PsuedoStyles map[string]map[string]map[string]string
}

func (s Styles) StyleTag(css string) {
	styleMaps := ParseCSS(css)

	if s.StyleMap == nil {
		s.StyleMap = map[string][]*StyleMap{}
	}

	for k, v := range styleMaps {
		if s.StyleMap[k] == nil {
			s.StyleMap[k] = []*StyleMap{}
		}
		s.StyleMap[k] = append(s.StyleMap[k], v...)
	}
}

var inheritedProps = []string{
	"color",
	"cursor",
	"font",
	"font-family",
	"font-style",
	"font-weight",
	"letter-spacing",
	"line-height",
	// "text-align",
	"text-indent",
	"text-justify",
	"text-shadow",
	"text-transform",
	"text-decoration",
	"visibility",
	"word-spacing",
	"display",
	"scrollbar-color",
}

func QuickStyles(n *Node) {
	// Inherit styles from parent
	if n.parent != nil {
		ps := n.parent.ComputedStyle
		for _, prop := range inheritedProps {
			if value, ok := ps[prop]; ok && value != "" {
				n.ComputedStyle[prop] = value
			}
		}
	}
}

func ConditionalStyleHandler(n *Node, newStyles map[string]string) {
	// Reset all styles to inital
	styles := map[string]string{}

	for k, v := range n.InitalStyles {
		if newStyles[k] == "" && n.ComputedStyle[k] != v {
			styles[k] = v
		}
		n.ComputedStyle[k] = v
	}
	// Then inherit the parents styles
	for _, k := range inheritedProps {
		if newStyles[k] != "" {
			n.ComputedStyle[k] = newStyles[k]
		}
	}

	// Then apply the conditional styles
	if n.hovered && n.ConditionalStyles[":hover"] != nil {
		for k, v := range n.ConditionalStyles[":hover"] {
			n.ComputedStyle[k] = v
			styles[k] = v
		}
	}
	if n.focused && n.ConditionalStyles[":focus"] != nil {
		for k, v := range n.ConditionalStyles[":focus"] {
			n.ComputedStyle[k] = v
			styles[k] = v
		}
	}

	if n.ConditionalStyles[":hover"] != nil || n.ConditionalStyles[":focus"] != nil {
		for _, v := range n.Children {
			ConditionalStyleHandler(v, styles)
		}
	}
}

// !ISSUE: GetStyles only needs to be ran if a new node is added, and the inital run, or a style tag innerHTML chanages
// + rest can be done with a modified QuickStyles
// + kinda see that note for a complete list

func (s Styles) GetStyles(n *Node) {
	if strings.Contains(n.Properties.Id, "head") {
		return
	}

	styles := make(map[string]string)
	pseudoStyles := make(map[string]map[string]string)
	conditionalStyles := make(map[string]map[string]string)

	// Inherit styles from parent
	if n.parent != nil {
		ps := n.parent.ComputedStyle
		for _, prop := range inheritedProps {
			if value, ok := ps[prop]; ok && value != "" {
				styles[prop] = value
			}
		}
	}

	baseSelectors := GenBaseElements(n)
	testedSelectors := map[string]bool{}

	// !DEVMAN: You need to pre-sort the selectors by their .Sheet field to create the
	// + cascading effect of CSS

	styleMaps := []*StyleMap{}
	for _, v := range baseSelectors {
		sm := s.StyleMap[v]
		styleMaps = append(styleMaps, sm...)
	}
	sort.Slice(styleMaps, func(i, j int) bool {
		return styleMaps[i].Sheet < styleMaps[j].Sheet
	})
	for _, m := range styleMaps {
		if ShouldTestSelector(n, m.Selector) {
			testedSelectors[m.Selector] = true
			match, isPseudo := TestSelector(n, m.Selector)
			if match {
				if isPseudo {
					pseudoSelector := "::" + strings.Split(m.Selector, "::")[1]
					if pseudoStyles[pseudoSelector] == nil {
						pseudoStyles[pseudoSelector] = map[string]string{}
					}
					for k, v := range *m.Styles {
						if v == "" {
							continue
						}
						if pseudoStyles[pseudoSelector] == nil {
							pseudoStyles[pseudoSelector] = map[string]string{}
						}
						pseudoStyles[pseudoSelector][k] = v
					}
				} else {
					for k, v := range *m.Styles {
						if v == "" {
							continue
						}
						styles[k] = v
					}
				}
			} else {
				if !n.hovered && !isPseudo {
					if strings.Contains(m.Selector, ":hover") {
						n.hovered = true

						match, _ = TestSelector(n, m.Selector)

						if match {
							conditionalStyles[":hover"] = map[string]string{}
							for k, v := range *m.Styles {
								if v == "" {
									continue
								}
								conditionalStyles[":hover"][k] = v
							}
						}

						n.hovered = false
					}
				}
				if !n.focused && !isPseudo {
					if strings.Contains(m.Selector, ":focus") {
						n.focused = true

						match, _ = TestSelector(n, m.Selector)

						if match {
							conditionalStyles[":focus"] = map[string]string{}
							for k, v := range *m.Styles {
								if v == "" {
									continue
								}
								conditionalStyles[":focus"][k] = v
							}
						}

						n.focused = false
					}
				}
			}
		}
	}

	// Parse inline styles
	inlineStyles := ParseStyleAttribute(n.GetAttribute("style"))
	for k, v := range inlineStyles {
		if v == "" {
			continue
		}
		styles[k] = v
	}

	// Handle z-index inheritance
	if n.parent != nil && styles["z-index"] == "" {
		parentZIndex := n.parent.ComputedStyle["z-index"]
		if parentZIndex != "" {
			z, _ := strconv.Atoi(parentZIndex)
			z += 1
			styles["z-index"] = strconv.Itoa(z)
		}
	}

	if n.tagName == "img" {
		styles["background-image"] = "url(\"" + n.src + "\")"
		styles["background-size"] = "100% 100%"
	}

	n.ComputedStyle = styles

	n.InitalStyles = map[string]string{}
	for k, v := range styles {
		n.InitalStyles[k] = v
	}
	n.ConditionalStyles = conditionalStyles

	s.PsuedoStyles[n.Properties.Id] = pseudoStyles
}
