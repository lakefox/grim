package ul

import (
	"grim/cstyle"
	"grim/element"
)

func Init() cstyle.Transformer {
	return cstyle.Transformer{
		Selector: func(n *element.Node, c *cstyle.CSS) bool {
			return n.TagName == "ul"
		},
		Handler: func(n *element.Node, c *cstyle.CSS) *element.Node {
			// The reason tN (temporary Node) is used, is because we have to go through the n.Children and it makes it hard to insert/remove the old one
			// its better to just replace it

			// !ISSUE: make stylable
			for i, v := range n.Children {
				if v.TagName != "li" {
					continue
				}
				dot := v.CreateElement("div")
				element.QuickStyles(&dot)
				dot.Style("background", "#000")
				dot.Style("border-radius", "100px")
				dot.Style("width", "5px")
				dot.Style("height", "5px")
				dot.Style("margin-right", "10px")

				content := v.CreateElement("div")
				element.QuickStyles(&content)
				content.InnerText = v.InnerText

				for k, v := range v.Styles() {
					content.Style(k, v)
				}

				content.Style("display", "block")
				v.Children = append(v.Children, &dot)
				v.Children = append(v.Children, &content)

				n.Children[i].Style("display", "flex")
				n.Children[i].Style("align-items", "center")

				n.Children[i] = v
			}
			return n
		},
	}
}
