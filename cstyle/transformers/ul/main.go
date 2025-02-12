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
			tN := n.CreateElement(n.TagName)
			tN.Properties.Id = n.Properties.Id
			for i, v := range n.Children {
				if v.TagName != "li" {
					tN.AppendChild(v)
					continue
				}
				li := n.CreateElement("li")
				dot := li.CreateElement("div")
				dot.Style("background", "#000")
				dot.Style("border-radius", "100px")
				dot.Style("width", "5px")
				dot.Style("height", "5px")
				dot.Style("margin-right", "10px")

				content := li.CreateElement("div")
				content.InnerText = v.InnerText

				for k, v := range v.Styles() {
					content.Style(k, v)
				}
				// content.CStyle = c.QuickStyles(&content)
				content.Style("display", "block")
				li.AppendChild(&dot)
				li.AppendChild(&content)
				li.Parent = n

				n.Children[i].Style("display", "flex")
				n.Children[i].Style("align-items", "center")
				// li.CStyle = c.QuickStyles(&li)

				tN.AppendChild(&li)
			}
			n.Children = tN.Children
			return n
		},
	}
}
