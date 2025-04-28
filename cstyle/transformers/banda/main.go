package banda

import (
	"grim/cstyle"
	"grim/element"
)

func Init() cstyle.Transformer {
	return cstyle.Transformer{
		Selector: func(n *element.Node, c *cstyle.CSS) bool {
			ps := n.StyleSheets.PsuedoStyles[n.Properties.Id]
			if ps["::before"] != nil || ps["::after"] != nil {
				return true
			} else {
				return false
			}
		},
		Handler: func(n *element.Node, c *cstyle.CSS) *element.Node {
			ps := n.StyleSheets.PsuedoStyles[n.Properties.Id]
			if ps["::before"] != nil {
				before := n.CreateElement("before")
				before.Parent = n
				before.ComputedStyle["display"] = "inline"

				for k, v := range ps["::before"] {
					before.ComputedStyle[k] = v
				}

				before.InnerText = ps["::before"]["content"][1 : len(ps["::before"]["content"])-1]

				if len(n.Children) == 0 {
					AppendChild(n, &before)
				} else {
					InsertBefore(n, &before, n.Children[0])
				}
			}

			if ps["::after"] != nil {
				after := n.CreateElement("after")
				after.Parent = n
				after.ComputedStyle["display"] = "inline"

				for k, v := range ps["::after"] {
					after.ComputedStyle[k] = v
				}

				after.InnerText = ps["::after"]["content"][1 : len(ps["::after"]["content"])-1]

				AppendChild(n, &after)
			}

			return n
		},
	}
}

func InsertBefore(n, c, tgt *element.Node) {
	c.Properties.Id = element.GenerateUniqueId(n, c.TagName())
	nodeIndex := -1
	for i, v := range n.Children {
		if v.Properties.Id == tgt.Properties.Id {
			nodeIndex = i
			break
		}
	}

	if nodeIndex > 0 {
		n.Children = append(n.Children[:nodeIndex], append([]*element.Node{c}, n.Children[nodeIndex:]...)...)
	} else {
		n.Children = append([]*element.Node{c}, n.Children...)
	}
}

func AppendChild(n, c *element.Node) {
	c.Properties.Id = element.GenerateUniqueId(n, c.TagName())
	n.Children = append(n.Children, c)
}
