package text

import (
	"grim/cstyle"
	"grim/element"
	"grim/utils"
	"html"
	"strings"
)

func Init() cstyle.Transformer {
	return cstyle.Transformer{
		Selector: func(n *element.Node) bool {
			if !utils.ChildrenHaveText(n) && len(strings.TrimSpace(n.InnerText)) > 0 {
				return true
			} else {
				return false
			}
		},
		Handler: func(n *element.Node, c *cstyle.CSS) *element.Node {
			if utils.IsParent(*n, "head") {
				return n
			}
			words := strings.Split(strings.TrimSpace(n.InnerText), " ")
			n.InnerText = ""
			if n.Style["display"] == "inline" {
				n.InnerText = DecodeHTMLEscapes(words[0])
				for i := 0; i < len(words)-1; i++ {
					// Add the words backwards because you are inserting adjacent to the parent
					a := (len(words) - 1) - i
					if len(strings.TrimSpace(words[a])) > 0 {
						el := n.CreateElement("notaspan")
						el.InnerText = DecodeHTMLEscapes(words[a])
						el.Parent = n

						el.Style = c.QuickStyles(&el)
						el.Style["display"] = "inline"
						// el.Style["margin-top"] = "10px"

						n.Parent.InsertAfter(&el, n)
					}
				}

			} else {
				for i := 0; i < len(words); i++ {
					if len(strings.TrimSpace(words[i])) > 0 {
						el := n.CreateElement("notaspan")
						el.InnerText = DecodeHTMLEscapes(words[i])
						el.Parent = n

						el.Style = c.QuickStyles(&el)
						el.Style["display"] = "inline"
						el.Style["font-size"] = "1em"
						// el.Style["margin-top"] = "10px"

						n.AppendChild(&el)
					}
				}
			}

			return n
		},
	}
}

func DecodeHTMLEscapes(input string) string {
	return html.UnescapeString(input)
}
