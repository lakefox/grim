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
				dot.ComputedStyle["background-color"] = "#000"
				dot.ComputedStyle["border-radius"] = "100px"
				dot.ComputedStyle["width"] = "5px"
				dot.ComputedStyle["height"] = "5px"
				dot.ComputedStyle["margin-right"] = "10px"

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
				v.ComputedStyle["display"] = "flex"
				v.ComputedStyle["align-items"] = "center"

				dot.Properties.Id = element.GenerateUniqueId(v, dot.TagName)
				// fmt.Println(dot.Properties.Id )
				// content.Properties.Id = element.GenerateUniqueId(v, content.TagName)
				n.Children[i] = v

			}

			return n
		},
	}
}
