package ul

import (
	"grim/cstyle"
	"grim/element"
)

func Init() cstyle.Transformer {
	return cstyle.Transformer{
		Selector: func(n *element.Node) bool {
			return n.TagName == "ul"
		},
		Handler: func(n *element.Node, c *cstyle.CSS) *element.Node {
			// The reason tN (temporary Node) is used, is because we have to go through the n.Children and it makes it hard to insert/remove the old one
			// its better to just replace it

			// !ISSUE: make stylable
			tN := n.CreateElement(n.TagName)
			for _, v := range n.Children {
				li := n.CreateElement("li")
				li.CStyle = v.CStyle
				dot := li.CreateElement("div")
				dot.CStyle["background"] = "#000"
				dot.CStyle["border-radius"] = "100px"
				dot.CStyle["width"] = "5px"
				dot.CStyle["height"] = "5px"
				dot.CStyle["margin-right"] = "10px"

				content := li.CreateElement("div")
				content.InnerText = v.InnerText
				content.CStyle = v.CStyle
				content.CStyle = c.QuickStyles(&content)
				content.CStyle["display"] = "block"
				li.AppendChild(&dot)
				li.AppendChild(&content)
				li.Parent = n

				li.CStyle["display"] = "flex"
				li.CStyle["align-items"] = "center"
				li.CStyle = c.QuickStyles(&li)

				tN.AppendChild(&li)
			}
			n.Children = tN.Children
			return n
		},
	}
}
