package banda

import (
	"grim/cstyle"
	"grim/element"
)

func Init() cstyle.Transformer {
	return cstyle.Transformer{
		Selector: func(n *element.Node, c *cstyle.CSS) bool {
			ps := c.PsuedoStyles[n.Properties.Id]
			if ps["::before"] != nil || ps["::after"] != nil {
				return true
			} else {
				return false
			}
		},
		Handler: func(n *element.Node, c *cstyle.CSS) *element.Node {
			ps := c.PsuedoStyles[n.Properties.Id]
			if ps["::before"] != nil {
				before := n.CreateElement("before")
				before.Parent = n
				before.Style("display", "inline")

				for k, v := range ps["::before"] {
					before.Style(k, v)
				}

				before.InnerText = ps["::before"]["content"][1 : len(ps["::before"]["content"])-1]

				if len(n.Children) == 0 {
					n.AppendChild(&before)
				} else {
					n.InsertBefore(&before, n.Children[0])
				}
			}

			if ps["::after"] != nil {
				after := n.CreateElement("after")
				after.Parent = n
				after.Style("display", "inline")

				for k, v := range ps["::after"] {
					after.Style(k, v)
				}

				after.InnerText = ps["::after"]["content"][1 : len(ps["::after"]["content"])-1]

				n.AppendChild(&after)
			}

			return n
		},
	}
}
