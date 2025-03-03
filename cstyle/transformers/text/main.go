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
			if len(strings.TrimSpace(n.InnerText)) > 0 && !element.ChildrenHaveText(n) {
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
			if n.ComputedStyle["display"] == "inline" {
				n.InnerText = DecodeHTMLEscapes(words[0])
				for i := 0; i < len(words)-1; i++ {
					// Add the words backwards because you are inserting adjacent to the parent
					a := (len(words) - 1) - i
					if len(strings.TrimSpace(words[a])) > 0 {
						el := n.CreateElement("text")
						el.InnerText = DecodeHTMLEscapes(words[a])

						el.Parent = n.Parent
						element.QuickStyles(&el)
						el.ComputedStyle["display"] = "inline"
						el.ComputedStyle["font-size"] = "1em"
						InsertAfter(n.Parent, &el, n)
					}
				}

			} else {
				for i := 0; i < len(words); i++ {
					if len(strings.TrimSpace(words[i])) > 0 {
						el := n.CreateElement("text")
						el.InnerText = DecodeHTMLEscapes(words[i])
						el.Parent = n
						element.QuickStyles(&el)
						el.ComputedStyle["display"] = "inline"
						el.ComputedStyle["font-size"] = "1em"
						AppendChild(n, &el)
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

func InsertAfter(n, c, tgt *element.Node) {
	c.Properties.Id = element.GenerateUniqueId(n, c.TagName)
	nodeIndex := -1
	for i, v := range n.Children {
		if v.Properties.Id == tgt.Properties.Id {
			nodeIndex = i
			break
		}
	}
	if nodeIndex > -1 {
		n.Children = append(n.Children[:nodeIndex+1], append([]*element.Node{c}, n.Children[nodeIndex+1:]...)...)
	} else {
		AppendChild(n, c)
	}
}

func AppendChild(n, c *element.Node) {
	c.Properties.Id = element.GenerateUniqueId(n, c.TagName)
	n.Children = append(n.Children, c)
}
