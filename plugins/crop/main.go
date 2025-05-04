package crop

import (
	"fmt"
	"grim"
	"grim/element"
	"grim/utils"
)

func Init() cstyle.Plugin {
	return cstyle.Plugin{
		Selector: func(n *element.Node, c *cstyle.CSS) bool {
			if n.ComputedStyle["overflow"] != "" || n.ComputedStyle["overflow-x"] != "" || n.ComputedStyle["overflow-y"] != "" {
				return true
			} else {
				return false
			}
		},
		Handler: func(n *element.Node, c *cstyle.CSS) {
			self := c.State[n.Properties.Id]

			scrollTop, scrollLeft := findScroll(n)
			fmt.Println(n.Properties.Id,scrollTop, scrollLeft)

			containerHeight := self.Height
			contentHeight := float32(self.ScrollHeight)

			containerWidth := self.Width
			contentWidth := float32(self.ScrollWidth)

			for _, v := range n.Children {
				if v.TagName() == "grim-track" && v.GetAttribute("direction") == "y" {
					if containerHeight < contentHeight {
						p := c.State[v.Properties.Id]
						p.Hidden = false
						c.State[v.Properties.Id] = p
						if len(v.Children) > 0 {
							p = c.State[v.Children[0].Properties.Id]
							p.Hidden = false

							p.Height = (containerHeight / contentHeight) * containerHeight

							p.Y = self.Y + float32(scrollTop)

							c.State[v.Children[0].Properties.Id] = p
						}
					} else {
						p := c.State[v.Properties.Id]
						p.Hidden = true
						c.State[v.Properties.Id] = p
						if len(v.Children) > 0 {
							p = c.State[v.Children[0].Properties.Id]
							p.Hidden = true
							c.State[v.Children[0].Properties.Id] = p
						}
						scrollTop = 0
					}
				}
				if v.TagName() == "grim-track" && v.GetAttribute("direction") == "x" {
					if containerWidth < contentWidth {
						p := c.State[v.Properties.Id]
						p.Hidden = false
						c.State[v.Properties.Id] = p

						if len(v.Children) > 0 {
							p = c.State[v.Children[0].Properties.Id]
							p.Hidden = false

							containerHeight -= utils.ConvertToPixels(v.ComputedStyle["height"], self.EM, self.Width)

							p.Width = (containerWidth / contentWidth) * containerWidth

							p.X = self.X + float32(scrollLeft)
						}
						c.State[v.Children[0].Properties.Id] = p
					} else {
						p := c.State[v.Properties.Id]
						p.Hidden = true
						c.State[v.Properties.Id] = p

						if len(v.Children) > 0 {
							p = c.State[v.Children[0].Properties.Id]
							p.Hidden = true
							c.State[v.Children[0].Properties.Id] = p
						}
						scrollLeft = 0
					}
				}
			}
			scrollTop = int((float32(scrollTop) / ((containerHeight / contentHeight) * containerHeight)) * containerHeight)
			scrollLeft = int((float32(scrollLeft) / ((containerWidth / contentWidth) * containerWidth)) * containerWidth)

			if n.ComputedStyle["overflow-y"] == "hidden" || n.ComputedStyle["overflow-y"] == "clip" {
				scrollTop = 0
			}

			if n.ComputedStyle["overflow-x"] == "hidden" || n.ComputedStyle["overflow-x"] == "clip" {
				scrollLeft = 0
			}

			for _, v := range n.Children {
				if v.ComputedStyle["position"] == "fixed" || v.TagName() == "grim-track" {
					continue
				}
				child := c.State[v.Properties.Id]
				if ((child.Y+child.Height)-float32(scrollTop) < (self.Y) || (child.Y-float32(scrollTop)) > self.Y+self.Height) ||
					((child.X+child.Width)-float32(scrollLeft) < (self.X) || (child.X-float32(scrollLeft)) > self.X+self.Width) {
					child.Hidden = true
					c.State[v.Properties.Id] = child
				} else {
					child.Hidden = false
					xCrop := 0
					yCrop := 0
					width := int(child.Width)
					height := int(child.Height)

					if child.Y-float32(scrollTop) < (self.Y) {
						yCrop = int((self.Y) - (child.Y - float32(scrollTop)))
						height = int(child.Height) - yCrop
					} else if (child.Y-float32(scrollTop))+child.Height > self.Y+self.Height {
						diff := ((child.Y - float32(scrollTop)) + child.Height) - (self.Y + self.Height)
						height = int(child.Height) - int(diff)
					}

					if child.X-float32(scrollLeft) < (self.X) {
						xCrop = int((self.X) - (child.X - float32(scrollLeft)))
						w := child.Width
						if self.Width < w-float32(xCrop) {
							w = self.Width + float32(scrollLeft) - child.X
						}
						width = int(w) - xCrop

					} else if (child.X-float32(scrollLeft))+child.Width > self.X+self.Width {
						diff := ((child.X - float32(scrollLeft)) + child.Width) - (self.X + self.Width)
						width = int(child.Width) - int(diff)
					}

					child.Crop = element.Crop{
						X:      xCrop,
						Y:      yCrop,
						Width:  width,
						Height: height,
					}
					c.State[v.Properties.Id] = child

					updateChildren(v, c, scrollTop, scrollLeft)
				}
			}
			c.State[n.Properties.Id] = self
		},
	}
}

func updateChildren(n *element.Node, c *grim.CSS, offsetY, offsetX int) {
	self := c.State[n.Properties.Id]
	self.X -= float32(offsetX)
	self.Y -= float32(offsetY)
	c.State[n.Properties.Id] = self
	for _, v := range n.Children {
		updateChildren(v, c, offsetY, offsetX)
	}
}

func findScroll(n *element.Node) (int, int) {
	left, top := n.GetScroll()
	if top != 0 || left != 0 {
		return top, left
	} else {
		for _, v := range n.Children {
			if v.ComputedStyle["overflow"] == "" && v.ComputedStyle["overflow-x"] == "" && v.ComputedStyle["overflow-y"] == "" {
				s, l := findScroll(v)
				if s != 0 || l != 0 {
					return s, l
				}
			}
		}
		return 0, 0
	}
}
