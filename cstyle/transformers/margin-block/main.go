package marginblock

import (
	"grim/cstyle"
	"grim/element"
	"strings"
)

func Init() cstyle.Transformer {
	return cstyle.Transformer{
		Selector: func(n *element.Node, c *cstyle.CSS) bool {
			return n.GetStyle("margin-block") != "" || n.GetStyle("margin-block-start") != "" || n.GetStyle("margin-block-end") != ""
		},
		Handler: func(n *element.Node, c *cstyle.CSS) *element.Node {

			writingMode := n.GetStyle("writing-mode")
			mb := parseMarginBlock(n.GetStyle("margin-block"))

			if n.GetStyle("margin-block-start") != "" {
				mb[0] = n.GetStyle("margin-block-start")
			}
			if n.GetStyle("margin-block-end") != "" {
				mb[1] = n.GetStyle("margin-block-end")
			}

			// !ISSUE: Not working line36 broke margin

			if writingMode == "vertical-lr" {
				n.SetStyle("margin-left", mb[0])
				n.SetStyle("margin-right", mb[1])
			} else if writingMode == "vertical-rl" {
				// !ISSUE: This will not move everything over
				// + link: https://developer.mozilla.org/en-US/docs/Web/CSS/margin-block
				n.SetStyle("margin-left", mb[1])
				n.SetStyle("margin-right", mb[0])
			} else if n.GetStyle("margin") == "" {
				if n.GetStyle("margin-top") == "" {
					n.SetStyle("margin-top", mb[0])
				}
				if n.GetStyle("margin-bottom") == "" {
					n.SetStyle("margin-bottom", mb[0])
				}
			}

			return n
		},
	}
}

func parseMarginBlock(s string) []string {
	split := strings.Split(s, " ")

	switch len(split) {
	case 2:
		return split
	case 1:
		return []string{split[0], split[0]}
	default:
		return []string{}
	}
}
