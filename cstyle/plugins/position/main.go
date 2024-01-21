package position

import (
	"gui/cstyle"
	"gui/element"
	"gui/utils"
)

func Init() cstyle.Plugin {
	return cstyle.Plugin{
		Styles: map[string]string{
			"position": "*",
		},
		Level: 0,
		Handler: func(n *element.Node) {
			styleMap := n.Style
			width, height := n.Properties.Width, n.Properties.Height
			x, y := n.Properties.X, n.Properties.Y

			var top, left, right, bottom bool = false, false, false, false

			if styleMap["position"] == "absolute" {
				base := GetPositionOffsetNode(n)
				if styleMap["top"] != "" {
					v, _ := utils.ConvertToPixels(styleMap["top"], float32(n.Properties.EM), n.Parent.Properties.Width)
					y = v + base.Properties.Y
					top = true
				}
				if styleMap["left"] != "" {
					v, _ := utils.ConvertToPixels(styleMap["left"], float32(n.Properties.EM), n.Parent.Properties.Width)
					x = v + base.Properties.X
					left = true
				}
				if styleMap["right"] != "" {
					v, _ := utils.ConvertToPixels(styleMap["right"], float32(n.Properties.EM), n.Parent.Properties.Width)
					x = (base.Properties.Width - width) - v
					right = true
				}
				if styleMap["bottom"] != "" {
					v, _ := utils.ConvertToPixels(styleMap["bottom"], float32(n.Properties.EM), n.Parent.Properties.Width)
					y = (base.Properties.Height - height) - v
					bottom = true
				}
			} else {
				for i, v := range n.Parent.Children {
					if v.Properties.Id == n.Properties.Id {
						if i-1 > 0 {
							sibling := n.Parent.Children[i-1]
							if styleMap["display"] == "inline" {
								if sibling.Style["display"] == "inline" {
									y = sibling.Properties.Y
								} else {
									y = sibling.Properties.Y + sibling.Properties.Height
								}
							} else {
								y = sibling.Properties.Y + sibling.Properties.Height
							}
						}
						break
					} else if styleMap["display"] != "inline" {
						y += v.Properties.Margin.Top + v.Properties.Margin.Bottom + v.Properties.Padding.Top + v.Properties.Padding.Bottom + v.Properties.Height
					}
				}
			}

			// Display modes need to be calculated here

			relPos := !top && !left && !right && !bottom

			if left || relPos {
				x += n.Properties.Margin.Left
			}
			if top || relPos {
				y += n.Properties.Margin.Top
			}
			if right {
				x -= n.Properties.Margin.Right
			}
			if bottom {
				y -= n.Properties.Margin.Bottom
			}

			n.Properties.X = x
			n.Properties.Y = y
			n.Properties.Width = width
			n.Properties.Height = height
		},
	}
}

func GetPositionOffsetNode(n *element.Node) *element.Node {
	pos := n.Style["position"]

	if pos == "relative" {
		return n
	} else {
		if n.Parent.Properties.Node != nil {
			return GetPositionOffsetNode(n.Parent)
		} else {
			return nil
		}
	}
}