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
				if len(n.Children) == 0 {
					n.AppendChild(&before)
				} else {
					n.InsertBefore(&before, n.Children[0])
				}

				before.ComputedStyle["display"] = "inline"

				for k, v := range ps["::before"] {
					before.ComputedStyle[k] = v
				}

				before.SetInnerText(ps["::before"]["content"][1 : len(ps["::before"]["content"])-1])

			}

			if ps["::after"] != nil {
				after := n.CreateElement("after")
				n.AppendChild(&after)

				after.ComputedStyle["display"] = "inline"

				for k, v := range ps["::after"] {
					after.ComputedStyle[k] = v
				}

				after.SetInnerText(ps["::after"]["content"][1 : len(ps["::after"]["content"])-1])

			}

			return n
		},
	}
}
