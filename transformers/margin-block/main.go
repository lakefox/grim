package marginblock

import (
	"grim"
	"grim/element"
	"strings"
)

func Init() grim.Transformer {
	return grim.Transformer{
		Selector: func(n *element.Node, c *grim.CSS) bool {
			return n.ComputedStyle["margin-block"] != "" || n.ComputedStyle["margin-block-start"] != "" || n.ComputedStyle["margin-block-end"] != ""
		},
		Handler: func(n *element.Node, c *grim.CSS) *element.Node {

			writingMode := n.ComputedStyle["writing-mode"]
			mb := parseMarginBlock(n.ComputedStyle["margin-block"])

			if n.ComputedStyle["margin-block-start"] != "" {
				mb[0] = n.ComputedStyle["margin-block-start"]
			}
			if n.ComputedStyle["margin-block-end"] != "" {
				mb[1] = n.ComputedStyle["margin-block-end"]
			}

			// !ISSUE: Not working line36 broke margin

			if writingMode == "vertical-lr" {
				n.ComputedStyle["margin-left"] = mb[0]
				n.ComputedStyle["margin-right"] = mb[1]
			} else if writingMode == "vertical-rl" {
				// !ISSUE: This will not move everything over
				// + link: https://developer.mozilla.org/en-US/docs/Web/CSS/margin-block
				n.ComputedStyle["margin-left"] = mb[1]
				n.ComputedStyle["margin-right"] = mb[0]
			} else if n.ComputedStyle["margin"] == "" {
				if n.ComputedStyle["margin-top"] == "" {
					n.ComputedStyle["margin-top"] = mb[0]
				}
				if n.ComputedStyle["margin-bottom"] == "" {
					n.ComputedStyle["margin-bottom"] = mb[0]
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
