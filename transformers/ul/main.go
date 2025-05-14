package ul

import (
	"grim"
)

func Init() grim.Transformer {
	return grim.Transformer{
		Selector: func(n *grim.Node, c *grim.CSS) bool {
			return n.TagName() == "ul"
		},
		Handler: func(n *grim.Node, c *grim.CSS) *grim.Node {
			// !TODO: make ul/ol stylable
			for i, v := range n.Children {
				if v.TagName() != "li" {
					continue
				}
				dot := v.CreateElement("div")
				v.AppendChild(&dot)
				v.ComputedStyle["display"] = "flex"
				v.ComputedStyle["align-items"] = "center"

				dot.ComputedStyle["background-color"] = "#000"
				dot.ComputedStyle["border-radius"] = "100px"
				dot.ComputedStyle["width"] = "5px"
				dot.ComputedStyle["height"] = "5px"
				dot.ComputedStyle["margin-right"] = "10px"

				n.Children[i] = v

			}

			return n
		},
	}
}
