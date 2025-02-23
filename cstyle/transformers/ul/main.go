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
			// !TODO: make ul/ol stylable
			for i, v := range n.Children {
				if v.TagName != "li" {
					continue
				}
				dot := v.CreateElement("div")
				dot.Parent = n
				element.QuickStyles(&dot)
				dot.Style("background-color", "#000")
				dot.Style("border-radius", "100px")
				dot.Style("width", "5px")
				dot.Style("height", "5px")
				dot.Style("margin-right", "10px")

				// content := v.CreateElement("div")
				// element.QuickStyles(&content)
				// content.InnerText = v.InnerText
				//
				// for k, v := range v.Styles() {
				// 	content.Style(k, v)
				// }
				//
				// content.Style("display", "block")
				v.Children = append(v.Children, &dot)
				// v.Children = append(v.Children, &content)
				// element.QuickStyles(v)
				v.Style("display", "flex")
				v.Style("align-items", "center")

				dot.Properties.Id = element.GenerateUniqueId(v, dot.TagName)
				// fmt.Println(dot.Properties.Id )
				// content.Properties.Id = element.GenerateUniqueId(v, content.TagName)
				n.Children[i] = v

			}

			return n
		},
	}
}
