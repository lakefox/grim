package crop

import (
	"grim/cstyle"
	"grim/element"
)

func Init() cstyle.Plugin {
	return cstyle.Plugin{
		Selector: func(n *element.Node) bool {
			if n.Style["overflow"] != "" || n.Style["overflow-x"] != "" || n.Style["overflow-y"] != "" {
				return true
			} else {
				return false
			}
		},
		Handler: func(n *element.Node, state *map[string]element.State, c *cstyle.CSS) {
			// !TODO: Needs to find crop bounds for X
			s := *state
			self := s[n.Properties.Id]

			scrollTop, scrollLeft := findScroll(n)

			containerHeight := self.Height
			contentHeight := float32(self.ScrollHeight)

			containerWidth := self.Width
			contentWidth := float32(self.ScrollWidth)

			for _, v := range n.Children {
				if v.TagName == "grim-track" && v.GetAttribute("direction") == "y" {
					if containerHeight < contentHeight {
						p := s[v.Children[0].Properties.Id]

						p.Height = (containerHeight / contentHeight) * containerHeight

						p.Y = self.Y + float32(scrollTop)

						(*state)[v.Children[0].Properties.Id] = p
					} else {
						p := s[v.Properties.Id]
						p.Hidden = true
						(*state)[v.Properties.Id] = p
						p = s[v.Children[0].Properties.Id]
						p.Hidden = true
						(*state)[v.Children[0].Properties.Id] = p
					}
				}
				if v.TagName == "grim-track" && v.GetAttribute("direction") == "x" {
					if containerWidth < contentWidth {
						// containerHeight -= utils.ConvertToPixels(v.Style["height"], self.EM, self.Width)
						p := s[v.Children[0].Properties.Id]

						p.Width = (containerWidth / contentWidth) * containerWidth

						p.X = self.X + float32(scrollLeft)

						(*state)[v.Children[0].Properties.Id] = p
					} else {
						p := s[v.Properties.Id]
						p.Hidden = true
						(*state)[v.Properties.Id] = p
						p = s[v.Children[0].Properties.Id]
						p.Hidden = true
						(*state)[v.Children[0].Properties.Id] = p
					}
				}
			}

			scrollTop = int((float32(scrollTop) / ((containerHeight / contentHeight) * containerHeight)) * containerHeight)
			scrollLeft = int((float32(scrollLeft) / ((containerWidth / contentWidth) * containerWidth)) * containerWidth)

			if n.Style["overflow-y"] == "hidden" || n.Style["overflow-y"] == "clip" {
				scrollTop = 0
			}

			if n.Style["overflow-x"] == "hidden" || n.Style["overflow-x"] == "clip" {
				scrollLeft = 0
			}

			for _, v := range n.Children {
				if v.Style["position"] == "fixed" || v.TagName == "grim-track" {
					continue
				}
				child := s[v.Properties.Id]
				if ((child.Y+child.Height)-float32(scrollTop) < (self.Y) || (child.Y-float32(scrollTop)) > self.Y+self.Height) ||
					((child.X+child.Width)-float32(scrollLeft) < (self.X) || (child.X-float32(scrollLeft)) > self.X+self.Width) {
					child.Hidden = true
					(*state)[v.Properties.Id] = child
				} else {
					child.Hidden = false
					xCrop := 0
					yCrop := 0
					width := int(child.Width)
					height := int(child.Height)

					// fmt.Println(scrollTop, scrollLeft)

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
						// if self.Width < child.Width-float32(scrollLeft-int(self.X+self.Margin.Left)) {
						// 	width = int(self.Width)
						// } else {
						// 	width = int(child.Width) - (scrollLeft - int(self.X+self.Margin.Left))
						// }
					} else if (child.X-float32(scrollLeft))+child.Width > self.X+self.Width {
						diff := ((child.X - float32(scrollLeft)) + child.Width) - (self.X + self.Width)
						width = int(child.Width) - int(diff)
					}

					// !ISSUE: Elements disappear when out of view during the resize, because the element is cropped to much
					child.Crop = element.Crop{
						X:      xCrop,
						Y:      yCrop,
						Width:  width,
						Height: height,
					}
					// fmt.Println(child.Crop)
					(*state)[v.Properties.Id] = child

					updateChildren(v, state, scrollTop, scrollLeft)
				}
			}
			(*state)[n.Properties.Id] = self
		},
	}
}

func updateChildren(n *element.Node, state *map[string]element.State, offsetY, offsetX int) {
	self := (*state)[n.Properties.Id]
	self.X -= float32(offsetX)
	self.Y -= float32(offsetY)
	(*state)[n.Properties.Id] = self
	for _, v := range n.Children {
		updateChildren(v, state, offsetY, offsetX)
	}
}

func findScroll(n *element.Node) (int, int) {
	if n.ScrollTop != 0 || n.ScrollLeft != 0 {
		return n.ScrollTop, n.ScrollLeft
	} else {
		for _, v := range n.Children {
			if v.Style["overflow"] == "" && v.Style["overflow-x"] == "" && v.Style["overflow-y"] == "" {
				s, l := findScroll(v)
				if s != 0 || l != 0 {
					return s, l
				}
			}
		}
		return 0, 0
	}
}
