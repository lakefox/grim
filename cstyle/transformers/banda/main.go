package banda

import (
	"grim/cstyle"
	"grim/element"
)

func Init() cstyle.Transformer {
	return cstyle.Transformer{
		Selector: func(n *element.Node) bool {
			if n.PseudoElements["::before"] != nil || n.PseudoElements["::after"] != nil {
				return true
			} else {
				return false
			}
		},
		Handler: func(n *element.Node, c *cstyle.CSS) *element.Node {
			if n.PseudoElements["::before"] != nil {
				before := n.CreateElement("before")
				before.Parent = n
				before.CStyle, _ = c.GetStyles(&before)
				before.CStyle["display"] = "inline"

				for k, v := range n.PseudoElements["::before"] {
					before.CStyle[k] = v
				}

				before.InnerText = n.PseudoElements["::before"]["content"][1 : len(n.PseudoElements["::before"]["content"])-1]

				if len(n.Children) == 0 {
					n.AppendChild(&before)
				} else {
					n.InsertBefore(&before, n.Children[0])
				}
			}

			if n.PseudoElements["::after"] != nil {
				after := n.CreateElement("after")
				after.Parent = n
				after.CStyle, _ = c.GetStyles(&after)
				after.CStyle["display"] = "inline"

				for k, v := range n.PseudoElements["::after"] {
					after.CStyle[k] = v
				}

				after.InnerText = n.PseudoElements["::after"]["content"][1 : len(n.PseudoElements["::after"]["content"])-1]

				n.AppendChild(&after)
			}

			return n
		},
	}
}
