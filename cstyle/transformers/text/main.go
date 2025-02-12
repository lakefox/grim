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
			if n.TagName == "text" && len(n.Children) == 0 {
				return true
			} else {
				return false
			}
		},
		Handler: func(n *element.Node, c *cstyle.CSS) *element.Node {
			if nonRenderTags[n.TagName] {
				return n
			}

			words := strings.Split(strings.TrimSpace(n.InnerText), " ")
			n.InnerText = ""
			if n.Style("display") == "inline" {
				n.InnerText = DecodeHTMLEscapes(words[0])
				for i := 0; i < len(words)-1; i++ {
					// Add the words backwards because you are inserting adjacent to the parent
					a := (len(words) - 1) - i
					if len(strings.TrimSpace(words[a])) > 0 {
						el := n.CreateElement("text")
						el.InnerText = DecodeHTMLEscapes(words[a])
						el.Parent = n

						qs := c.QuickStyles(&el)

						for k, v := range qs {
							el.Style(k, v)
						}
						el.Style("display", "inline")
						// el.Style["margin-top"] = "10px"
						n.Parent.InsertAfter(&el, n)
					}
				}

			} else {
				for i := 0; i < len(words); i++ {
					if len(strings.TrimSpace(words[i])) > 0 {
						el := n.CreateElement("text")
						el.InnerText = DecodeHTMLEscapes(words[i])
						el.Parent = n

						qs := c.QuickStyles(&el)

						for k, v := range qs {
							el.Style(k, v)
						}
						el.Style("display", "inline")
						el.Style("font-size", "1em")
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
