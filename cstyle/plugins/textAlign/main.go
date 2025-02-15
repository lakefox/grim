package textAlign

import (
	// "fmt"
	"grim/cstyle"
	"grim/element"
)

func Init() cstyle.Plugin {
	return cstyle.Plugin{
		Selector: func(n *element.Node, c *cstyle.CSS) bool {
			return n.Style("text-align") != ""
		},
		Handler: func(n *element.Node, c *cstyle.CSS) {
			self := c.State[n.Properties.Id]
			minX := float32(9e15)
			maxXW := float32(0)

			nChildren := []*element.Node{}
			for _, v := range n.Children {
				// This prevents using absolutely positionioned elements in the alignment of text
				// + Will need to add the other styles
				if v.Style("position") != "absolute" {
					nChildren = append(nChildren, v)
				}
			}


			align := n.Style("text-align")
			// fmt.Println(n.Properties.Id, align, nChildren, n.Children)
			if align == "center" {
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
								cState.X += self.Padding.Left + ((((self.Width - (self.Padding.Left + self.Padding.Right)) + (self.Border.Left.Width + self.Border.Right.Width)) - (maxXW - minX)) / 2) - (minX - self.X)

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
						// fmt.Println("1",a,n.Properties.Id, nChildren[a].Properties.Id, cState.X)
						cState.X += self.Padding.Left + ((((self.Width - (self.Padding.Left + self.Padding.Right)) + (self.Border.Left.Width + self.Border.Right.Width)) - (maxXW - minX)) / 2) - (minX - self.X)

						c.State[nChildren[a].Properties.Id] = cState
						// fmt.Println("2",a,n.Properties.Id, nChildren[a].Properties.Id, cState.X)
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
