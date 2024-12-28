package crop

import (
	"grim/cstyle"
	"grim/element"
	"strings"
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
		Level: 1,
		Handler: func(n *element.Node, state *map[string]element.State) {
			// !TODO: Needs to find crop bounds for X
			s := *state
			self := s[n.Properties.Id]

			scrollTop := findScroll(n)

			containerHeight := self.Height
			contentHeight := float32(self.ScrollHeight)

			for _, v := range n.Children {
				if v.TagName == "grim-track" {
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
					break
				}
			}

			scrollTop = int((float32(scrollTop) / ((containerHeight / contentHeight) * containerHeight)) * containerHeight)

			if containerHeight > contentHeight {
				return
			}

			if n.Style["overflow-y"] == "hidden" || n.Style["overflow-y"] == "clip" {
				scrollTop = 0
			}

			if n.Style["overflow-y"] == "visible" {
				return
			}

			for _, v := range n.Children {
				if v.Style["position"] == "fixed" || v.TagName == "grim-track" {
					continue
				}
				child := s[v.Properties.Id]

				if (child.Y+child.Height)-float32(scrollTop) < (self.Y) || (child.Y-float32(scrollTop)) > self.Y+self.Height {
					child.Hidden = true
					(*state)[v.Properties.Id] = child
				} else {
					child.Hidden = false
					yCrop := 0
					height := int(child.Height)
					if child.Y-float32(scrollTop) < (self.Y) {
						yCrop = int((self.Y) - (child.Y - float32(scrollTop)))
						height = int(child.Height) - yCrop
					} else if (child.Y-float32(scrollTop))+child.Height > self.Y+self.Height {
						diff := ((child.Y - float32(scrollTop)) + child.Height) - (self.Y + self.Height)
						height = int(child.Height) - int(diff)
					}
					// !ISSUE: Elements disappear when out of view during the resize, because the element is cropped to much
					child.Crop = element.Crop{
						X:      0,
						Y:      yCrop,
						Width:  int(child.Width),
						Height: height,
					}
					(*state)[v.Properties.Id] = child

					updateChildren(v, state, scrollTop)
				}
			}
			(*state)[n.Properties.Id] = self
		},
	}
}

func updateChildren(n *element.Node, state *map[string]element.State, offset int) {
	self := (*state)[n.Properties.Id]
	self.Y -= float32(offset)
	(*state)[n.Properties.Id] = self
	for _, v := range n.Children {
		updateChildren(v, state, offset)
	}
}

func findBounds(n *element.Node, state *map[string]element.State) (float32, float32) {
	s := *state
	var minY, maxY float32
	minY = 10e10
	for _, v := range n.Children {
		if strings.HasPrefix(v.TagName, "grim") {
			continue
		}
		child := s[v.Properties.Id]
		if child.Y < minY {
			minY = child.Y
		}
		if child.Y+child.Height > maxY {
			maxY = child.Y + child.Height
		}
		nMinY, nMaxY := findBounds(v, state)

		if nMinY < minY {
			minY = nMinY
		}
		if nMaxY > maxY {
			maxY = nMaxY
		}
	}
	return minY, maxY
}

func findScroll(n *element.Node) int {
	if n.ScrollTop != 0 {
		return n.ScrollTop
	} else {
		for _, v := range n.Children {
			if v.Style["overflow"] == "" && v.Style["overflow-x"] == "" && v.Style["overflow-y"] == "" {
				s := findScroll(v)
				if s != 0 {
					return s
				}
			}
		}
		return 0
	}
}
