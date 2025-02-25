package element

import (
	"sort"
	"strconv"
	"strings"
)

type Styles struct {
	StyleMap     map[string][]*StyleMap
	PsuedoStyles map[string]map[string]map[string]string
	Styles       map[string]string
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
	styles := make(map[string]string)

	// Inherit styles from parent
	if n.Parent != nil {
		ps := n.Parent.Styles()
		for _, prop := range inheritedProps {
			if value, ok := ps[prop]; ok && value != "" {
				styles[prop] = value
			}
		}
	}

	// Add node's own styles
	for k, v := range styles {
		n.Style(k, v)
	}
}

// !ISSUE: GetStyles only needs to be ran if a new node is added, and the inital run, or a style tag innerHTML chanages
// + rest can be done with a modified QuickStyles
// + kinda see that note for a complete list

func (s Styles) GetStyles(n *Node) {
	if strings.Contains(n.Properties.Id, "head") {
		return
	}
	// fmt.Println(n.Properties.Id)
	styles := make(map[string]string)
	pseudoStyles := make(map[string]map[string]string)

	// Inherit styles from parent
	if n.Parent != nil {
		ps := n.Parent.Styles()
		for _, prop := range inheritedProps {
			if value, ok := ps[prop]; ok && value != "" {
				styles[prop] = value
			}
		}
	}

	// Add node's own styles
	for k, v := range n.Styles() {
		styles[k] = v
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
						if pseudoStyles[pseudoSelector] == nil {
							pseudoStyles[pseudoSelector] = map[string]string{}
						}
						pseudoStyles[pseudoSelector][k] = v
					}
				} else {
					for k, v := range *m.Styles {
						styles[k] = v
					}
				}
			}
		}
	}

	// Parse inline styles
	inlineStyles := ParseStyleAttribute(n.GetAttribute("style"))
	for k, v := range inlineStyles {
		styles[k] = v
	}

	// Handle z-index inheritance
	if n.Parent != nil && styles["z-index"] == "" {
		parentZIndex := n.Parent.Style("z-index")
		if parentZIndex != "" {
			z, _ := strconv.Atoi(parentZIndex)
			z += 1
			styles["z-index"] = strconv.Itoa(z)
		}
	}
	if n.StyleSheets.Styles == nil {
		n.StyleSheets.Styles = map[string]string{}
	}
	for k, v := range styles {
		n.Style(k, v)
		n.StyleSheets.Styles[k] = v
	}
	s.PsuedoStyles[n.Properties.Id] = pseudoStyles
}
