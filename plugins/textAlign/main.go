package textAlign

import (
	"grim"
)

func Init() grim.Plugin {
	return grim.Plugin{
		Selector: func(n *grim.Node, c *grim.CSS) bool {
			return n.ComputedStyle["text-align"] != ""
		},
		Handler: func(n *grim.Node, c *grim.CSS) {
			self := c.State[n.Properties.Id]
			minX := float32(9e15)
			maxXW := float32(0)

			nChildren := []*grim.Node{}
			for _, v := range n.Children {
				// This prevents using absolutely positionioned elements in the alignment of text
				// + Will need to add the other styles
				if v.ComputedStyle["position"] != "absolute" {
					nChildren = append(nChildren, v)
				}
			}

			align := n.ComputedStyle["text-align"]

			if align == "center" {
				if len(nChildren) > 0 {
					minX = c.State[nChildren[0].Properties.Id].X
					baseY := c.State[nChildren[0].Properties.Id].Y + c.State[nChildren[0].Properties.Id].Height
					last := 0
					for i := 0; i < len(nChildren)-1; i++ {
						cState := c.State[nChildren[i].Properties.Id]
						next := c.State[nChildren[i+1].Properties.Id]

						if cState.X < minX {
							minX = cState.X
						}
						if (cState.X + cState.Width) > maxXW {
							maxXW = cState.X + cState.Width
						}

						if baseY != next.Y+next.Height {
							baseY = next.Y + next.Height
							for a := last; a < i+1; a++ {
								cState := c.State[nChildren[a].Properties.Id]
								cState.X += self.Padding.Left + ((((self.Width - (self.Padding.Left + self.Padding.Right)) + (self.Border.Left.Width + self.Border.Right.Width)) - (maxXW - minX)) / 2) - (minX - self.X)

								if cState.X < self.X+self.Padding.Left {
									cState.X = self.X + self.Padding.Left
								}
								c.State[nChildren[a].Properties.Id] = cState
							}
							minX = 9e15
							maxXW = 0
							last = i + 1
						}

					}
					minX = c.State[nChildren[last].Properties.Id].X
					maxXW = c.State[nChildren[len(nChildren)-1].Properties.Id].X + c.State[nChildren[len(nChildren)-1].Properties.Id].Width
					for a := last; a < len(nChildren); a++ {
						cState := c.State[nChildren[a].Properties.Id]
						cState.X += self.Padding.Left + ((((self.Width - (self.Padding.Left + self.Padding.Right)) + (self.Border.Left.Width + self.Border.Right.Width)) - (maxXW - minX)) / 2) - (minX - self.X)

						if cState.X < self.X+self.Padding.Left {
							cState.X = self.X + self.Padding.Left
						}
						c.State[nChildren[a].Properties.Id] = cState
					}
				}
			} else if align == "right" {
				if len(nChildren) > 0 {
					minX = c.State[nChildren[0].Properties.Id].X
					baseY := c.State[nChildren[0].Properties.Id].Y + c.State[nChildren[0].Properties.Id].Height
					last := 0
					for i := 1; i < len(nChildren)-1; i++ {

						cState := c.State[nChildren[i].Properties.Id]
						next := c.State[nChildren[i+1].Properties.Id]

						if cState.X < minX {
							minX = cState.X
						}
						if (cState.X + cState.Width) > maxXW {
							maxXW = cState.X + cState.Width
						}

						if baseY != next.Y+next.Height {
							baseY = next.Y + next.Height
							for a := last; a < i+1; a++ {
								cState := c.State[nChildren[a].Properties.Id]
								cState.X += ((self.Width + (self.Border.Left.Width + self.Border.Right.Width)) - (maxXW - minX)) + ((self.X - minX) * 2)
								c.State[nChildren[a].Properties.Id] = cState
							}
							minX = 9e15
							maxXW = 0
							last = i + 1
						}

					}
					minX = c.State[nChildren[last].Properties.Id].X
					maxXW = c.State[nChildren[len(nChildren)-1].Properties.Id].X + c.State[nChildren[len(nChildren)-1].Properties.Id].Width
					for a := last; a < len(nChildren); a++ {
						cState := c.State[nChildren[a].Properties.Id]
						cState.X += ((self.Width + (self.Border.Left.Width + self.Border.Right.Width)) - (maxXW - minX)) + ((self.X - minX) * 2)
						c.State[nChildren[a].Properties.Id] = cState
					}

				}
			}

			c.State[n.Properties.Id] = self
		},
	}
}
