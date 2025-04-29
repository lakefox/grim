package text

import (
	"grim/cstyle"
	"grim/element"
	"html"
	"strings"
)

var nonRenderTags = map[string]bool{
	"head":  true,
	"meta":  true,
	"link":  true,
	"title": true,
	"style": true,
}

func Init() cstyle.Transformer {
	return cstyle.Transformer{
		Selector: func(n *element.Node, c *cstyle.CSS) bool {
			if len(strings.TrimSpace(n.GetInnerText())) > 0 && !element.ChildrenHaveText(n) {
				return true
			} else {
				return false
			}
		},
		Handler: func(n *element.Node, c *cstyle.CSS) *element.Node {
			if nonRenderTags[n.TagName()] {
				return n
			}
			words := strings.Split(strings.TrimSpace(DecodeHTMLEscapes(n.GetInnerText())), " ")
			n.SetInnerText("")
			if n.ComputedStyle["display"] == "inline" {
				n.SetInnerText(DecodeHTMLEscapes(words[0]))
				for i := 0; i < len(words)-1; i++ {
					// Add the words backwards because you are inserting adjacent to the parent
					a := (len(words) - 1) - i
					if len(strings.TrimSpace(words[a])) > 0 {
						el := n.CreateElement("text")
						el.SetInnerText(words[a])
						// !CHECK: Idk if this breaks text
						// el.Parent = n
						// element.QuickStyles(&el)
						el.ComputedStyle["display"] = "inline"
						el.ComputedStyle["font-size"] = "1em"
						n.Parent().InsertAfter(&el, n)
					}
				}

			} else {
				for i := 0; i < len(words); i++ {
					if len(strings.TrimSpace(words[i])) > 0 {
						el := n.CreateElement("text")
						el.SetInnerText(words[i])
						// el.Parent = n
						// element.QuickStyles(&el)
						el.ComputedStyle["display"] = "inline"
						el.ComputedStyle["font-size"] = "1em"
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
