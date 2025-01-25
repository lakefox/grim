package marginblock

import (
	"grim/cstyle"
	"grim/element"
	"strings"
)

func Init() cstyle.Transformer {
	return cstyle.Transformer{
		Selector: func(n *element.Node) bool {
			return n.CStyle["margin-block"] != "" || n.CStyle["margin-block-start"] != "" || n.CStyle["margin-block-end"] != ""
		},
		Handler: func(n *element.Node, c *cstyle.CSS) *element.Node {

			writingMode := n.CStyle["writing-mode"]
			mb := parseMarginBlock(n.CStyle["margin-block"])

			if n.CStyle["margin-block-start"] != "" {
				mb[0] = n.CStyle["margin-block-start"]
			}
			if n.CStyle["margin-block-end"] != "" {
				mb[1] = n.CStyle["margin-block-end"]
			}

			// !ISSUE: Not working tline36 broke margin

			if writingMode == "vertical-lr" {
				n.CStyle["margin-left"] = mb[0]
				n.CStyle["margin-right"] = mb[1]
			} else if writingMode == "vertical-rl" {
				// !ISSUE: This will not move everything over
				// + link: https://developer.mozilla.org/en-US/docs/Web/CSS/margin-block
				n.CStyle["margin-left"] = mb[1]
				n.CStyle["margin-right"] = mb[0]
			} else if n.CStyle["margin"] == "" {
				if n.CStyle["margin-top"] == "" {
					n.CStyle["margin-top"] = mb[0]
				}
				if n.CStyle["margin-bottom"] == "" {
					n.CStyle["margin-bottom"] = mb[0]
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
