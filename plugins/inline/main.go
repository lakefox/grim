package inline

import (
	"grim"
	"math"
)

func Init() grim.Plugin {
	return grim.Plugin{
		Selector: func(n *grim.Node, c *grim.CSS) bool {
			return n.ComputedStyle["display"] == "inline"
		},
		Handler: func(n *grim.Node, c *grim.CSS) {
			self := c.State[n.Properties.Id]
			p := n.Parent()
			parent := c.State[p.Properties.Id]
			copyOfX := self.X
			copyOfY := self.Y
			if copyOfX < parent.X+parent.Padding.Left {
				copyOfX = parent.X + parent.Padding.Left
			}

			for i, v := range p.Children {
				if i > 0 {
					if v.ComputedStyle["position"] != "absolute" {
						if v.Properties.Id == n.Properties.Id {
							sib := p.Children[i-1]
							if sib.ComputedStyle["position"] == "absolute" {
								if i-2 >= 0 {
									sib = p.Children[i-2]
								} else {
									continue
								}
							}
							sibling := c.State[sib.Properties.Id]
							if sibling.X+sibling.Width+self.Width > (parent.Width+parent.X+parent.Border.Left.Width)-(parent.Padding.Right) {
								// Break Node.Id
								self.Y = sibling.Y + sibling.Height
								self.X = copyOfX
							} else {
								// Node did not break
								if sib.ComputedStyle["display"] != "inline" {
									self.Y = sibling.Y + sibling.Height + sibling.Border.Top.Width + sibling.Border.Bottom.Width + sibling.Margin.Top + sibling.Margin.Bottom
								} else {
									self.Y = sibling.Y
									self.X = sibling.X + sibling.Width
								}
								if n.InnerText() != "" {
									baseY := sibling.Y
									var max float32
									for a := i; a >= 0; a-- {
										b := p.Children[a]
										bStyle := c.State[b.Properties.Id]
										if bStyle.Y == baseY {
											if bStyle.EM > max {
												max = bStyle.EM
											}
										}
									}

									for a := i; a >= 0; a-- {
										b := p.Children[a]
										bStyle := c.State[b.Properties.Id]
										if bStyle.Y == baseY {
											bStyle.Y += (float32(math.Ceil(float64((max - (max * 0.3))))) - float32(math.Ceil(float64(bStyle.EM-(bStyle.EM*0.3)))))
											c.State[b.Properties.Id] = bStyle
										}
									}
									if self.Y == baseY {
										self.Y += (float32(math.Ceil(float64((max - (max * 0.3))))) - float32(math.Ceil(float64(self.EM-(self.EM*0.3)))))
									}
								}

							}
							break
						}
					}
				}

			}
			propagateOffsets(n, copyOfX, copyOfY, self, c)
			c.State[n.Properties.Id] = self
		},
	}
}

func propagateOffsets(n *grim.Node, copyOfX, copyOfY float32, self grim.State, c *grim.CSS) {
	for _, v := range n.Children {
		vState := c.State[v.Properties.Id]
		vState.X += self.X - copyOfX
		vState.Y += self.Y - copyOfY
		if len(v.Children) > 0 {
			propagateOffsets(v, copyOfX, copyOfY, self, c)
		}
		c.State[v.Properties.Id] = vState
	}
}
