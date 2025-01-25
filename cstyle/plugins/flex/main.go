package flex

import (
	"grim/cstyle"
	"grim/cstyle/plugins/inline"
	"grim/element"
	"grim/utils"
	"sort"
	"strings"
)

func Init() cstyle.Plugin {
	return cstyle.Plugin{
		Selector: func(n *element.Node) bool {
			styles := map[string]string{
				"display": "flex",
			}
			matches := true
			for name, value := range styles {
				if (n.CStyle[name] != value || n.CStyle[name] == "") && !(value == "*") {
					matches = false
				}
			}

			return matches
		},
		Handler: func(n *element.Node, state *map[string]element.State, c *cstyle.CSS) {
			s := *state
			self := s[n.Properties.Id]

			verbs := strings.Split(n.CStyle["flex-direction"], "-")
			flexDirection := verbs[0]
			if flexDirection == "" {
				flexDirection = "row"
			}
			flexReversed := false
			if len(verbs) > 1 {
				flexReversed = true
			}

			var flexWrapped bool
			if n.CStyle["flex-wrap"] == "wrap" {
				flexWrapped = true
			} else {
				flexWrapped = false
			}

			alignContent := n.CStyle["align-content"]
			if alignContent == "" {
				alignContent = "normal"
			}
			alignItems := n.CStyle["align-items"]
			if alignItems == "" {
				alignItems = "normal"
			}
			justifyItems := n.CStyle["justify-items"]
			if justifyItems == "" {
				justifyItems = "normal"
			}

			justifyContent := n.CStyle["justify-content"]
			if justifyContent == "" {
				justifyContent = "normal"
			}
			rows := [][]int{}
			maxH := float32(0)
			// maxW := float32(0)

			// Get inital sizing
			textTotal := 0
			textCounts := []int{}
			widths := []float32{}
			// heights := []float32{}
			innerSizes := [][]float32{}
			minWidths := []float32{}
			minHeights := []float32{}
			maxWidths := []float32{}
			// maxHeights := []float32{}
			for _, v := range n.Children {
				count := countText(v)
				textTotal += count
				textCounts = append(textCounts, count)

				minw := getMinWidth(v, state)
				minWidths = append(minWidths, minw)

				maxw := getMaxWidth(v, state)
				maxWidths = append(maxWidths, maxw)

				w, h := getInnerSize(v, state)

				minh := getMinHeight(v, state)
				minHeights = append(minHeights, minh)

				// maxh := getMaxHeight(&v, state)
				// maxHeights = append(maxHeights, maxh)
				innerSizes = append(innerSizes, []float32{w, h})
			}
			selfWidth := (self.Width - self.Padding.Left) - self.Padding.Right
			selfHeight := (self.Height - self.Padding.Top) - self.Padding.Bottom

			if flexDirection == "row" {
				// if the elements are less than the size of the parent, don't change widths. Just set mins
				if !flexWrapped {
					if add2d(innerSizes, 0) < selfWidth {
						for i := range innerSizes {
							// for i, _ := range n.Children {
							// vState := s[v.Properties.Id]

							w := innerSizes[i][0]
							// w -= vState.Margin.Left + vState.Margin.Right + (vState.Border.Width * 2)
							widths = append(widths, w)
						}
					} else {
						// Modifiy the widths so they aren't under the mins
						for i, v := range n.Children {
							vState := s[v.Properties.Id]

							w := ((selfWidth / float32(textTotal)) * float32(textCounts[i]))
							w -= vState.Margin.Left + vState.Margin.Right + (vState.Border.Left.Width + vState.Border.Right.Width)

							if w < minWidths[i] {
								selfWidth -= minWidths[i] + vState.Margin.Left + vState.Margin.Right + (vState.Border.Left.Width + vState.Border.Right.Width)
								textTotal -= textCounts[i]
								textCounts[i] = 0
							}

						}
						for i, v := range n.Children {
							vState := s[v.Properties.Id]

							w := ((selfWidth / float32(textTotal)) * float32(textCounts[i]))
							w -= vState.Margin.Left + vState.Margin.Right + (vState.Border.Left.Width + vState.Border.Right.Width)
							// (w!=w) is of NaN
							if w < minWidths[i] || (w != w) {
								w = minWidths[i]
							}
							widths = append(widths, w)
						}
					}
					// Apply the new widths
					fState := s[n.Children[0].Properties.Id]
					for i, v := range n.Children {
						vState := s[v.Properties.Id]

						vState.Width = widths[i]
						xStore := vState.X
						if i > 0 {
							sState := s[n.Children[i-1].Properties.Id]
							vState.X = sState.X + sState.Width + sState.Margin.Right + vState.Margin.Left + vState.Border.Left.Width + vState.Border.Right.Width
							propagateOffsets(v, xStore, vState.Y, vState.X, fState.Y+vState.Margin.Top, state)
						}

						vState.Y = fState.Y + vState.Margin.Top

						(*state)[v.Properties.Id] = vState
						deInline(v, state)
						applyInline(v, state, c)
						applyBlock(v, state)
						_, h := getInnerSize(v, state)
						h = utils.Max(h, vState.Height)
						maxH = utils.Max(maxH, h)
					}
					// When not wrapping everything will be on the same row
					rows = append(rows, []int{0, len(n.Children), int(maxH)})
				} else {
					// Flex Wrapped
					sum := innerSizes[0][0]
					for i := 0; i < len(n.Children); i++ {
						v := n.Children[i]
						vState := s[v.Properties.Id]

						// if the next plus current will break then
						w := innerSizes[i][0]
						if i > 0 {
							sib := s[n.Children[i-1].Properties.Id]
							if maxWidths[i] > selfWidth {
								w = selfWidth - vState.Margin.Left - vState.Margin.Right - (vState.Border.Left.Width + vState.Border.Right.Width)
							}
							if w+sum > selfWidth {
								sum = w + vState.Margin.Left + vState.Margin.Right + (vState.Border.Left.Width + vState.Border.Right.Width)
							} else {
								propagateOffsets(v, vState.X, vState.Y, vState.X, sib.Y, state)
								vState.Y = sib.Y
								(*state)[v.Properties.Id] = vState
								sum += w + vState.Margin.Left + vState.Margin.Right + (vState.Border.Left.Width + vState.Border.Right.Width)
							}
						}

						widths = append(widths, w)
					}

					// Move the elements into the correct position
					start := 0
					var prevOffset float32
					for i := 0; i < len(n.Children); i++ {
						v := n.Children[i]
						vState := s[v.Properties.Id]

						vState.Width = widths[i]
						xStore := vState.X
						yStore := vState.Y

						if i > 0 {
							sib := s[n.Children[i-1].Properties.Id]
							if vState.Y+prevOffset == sib.Y {
								yStore += prevOffset

								if vState.Height < sib.Height {
									vState.Height = minHeight(v, state, sib.Height)
								}
								// Shift right if on a row with sibling
								xStore = sib.X + sib.Width + sib.Margin.Right + sib.Border.Left.Width + vState.Margin.Left + vState.Border.Left.Width
							} else {
								// Shift under sibling
								yStore = sib.Y + sib.Height + sib.Margin.Top + sib.Margin.Bottom + sib.Border.Top.Width + sib.Border.Bottom.Width
								prevOffset = yStore - vState.Y
								rows = append(rows, []int{start, i, int(maxH)})
								start = i
								maxH = 0
							}
							propagateOffsets(v, vState.X, vState.Y, xStore, yStore, state)
						}
						vState.X = xStore
						vState.Y = yStore

						(*state)[v.Properties.Id] = vState
						deInline(v, state)
						applyInline(v, state, c)
						applyBlock(v, state)
						_, h := getInnerSize(v, state)
						h = utils.Max(h, vState.Height)
						maxH = utils.Max(maxH, h)
						vState.Height = minHeight(v, state, h)
						(*state)[v.Properties.Id] = vState
					}
					if start < len(n.Children) {
						rows = append(rows, []int{start, len(n.Children), int(maxH)})
					}
				}

				var totalHeight float32

				for _, v := range rows {
					totalHeight += float32(v[2])
					for i := v[0]; i < v[1]; i++ {
						vState := s[n.Children[i].Properties.Id]
						if vState.Height == float32(v[2]) {
							totalHeight += vState.Margin.Top + vState.Margin.Bottom
							v[2] += int(vState.Margin.Top + vState.Margin.Bottom + (vState.Border.Top.Width + vState.Border.Bottom.Width))
						}
					}
				}
				var yOffset float32
				for _, v := range rows {
					for i := v[0]; i < v[1]; i++ {
						vState := s[n.Children[i].Properties.Id]
						// height := float32(v[2])
						if n.Children[i].CStyle["height"] == "" && n.Children[i].CStyle["min-height"] == "" && !flexWrapped {
							height := self.Height - self.Padding.Top - self.Padding.Bottom - vState.Margin.Top - vState.Margin.Bottom - (vState.Border.Top.Width + vState.Border.Bottom.Width)
							vState.Height = minHeight(n.Children[i], state, height)
						} else if flexWrapped && (n.CStyle["height"] != "" || n.CStyle["min-height"] != "") {
							height := ((selfHeight / totalHeight) * float32(v[2])) - (vState.Margin.Top + vState.Margin.Bottom + (vState.Border.Top.Width + vState.Border.Bottom.Width))
							vState.Height = minHeight(n.Children[i], state, height)
							yStore := vState.Y
							vState.Y = self.Y + self.Padding.Top + yOffset + vState.Margin.Top + vState.Border.Top.Width
							propagateOffsets(n.Children[i], vState.X, yStore, vState.X, vState.Y, state)
						} else if flexWrapped {
							if vState.Height+vState.Margin.Top+vState.Margin.Bottom+(vState.Border.Top.Width+vState.Border.Bottom.Width) != float32(v[2]) {
								height := vState.Height - (vState.Margin.Top + vState.Margin.Bottom + (vState.Border.Top.Width + vState.Border.Bottom.Width))
								vState.Height = minHeight(n.Children[i], state, height)
							}
							yStore := vState.Y
							vState.Y = self.Y + self.Padding.Top + yOffset + vState.Margin.Top + vState.Border.Top.Width
							propagateOffsets(n.Children[i], vState.X, yStore, vState.X, vState.Y, state)
						}
						(*state)[n.Children[i].Properties.Id] = vState
					}
					if flexWrapped && (n.CStyle["height"] != "" || n.CStyle["min-height"] != "") {
						yOffset += ((selfHeight / totalHeight) * float32(v[2]))
					} else if flexWrapped {
						yOffset += float32(v[2])
					}
				}
				// Reverse elements
				if flexReversed {
					rowReverse(rows, n, state)
				}

				if justifyContent != "" && justifyContent != "normal" {
					justifyRow(rows, n, state, justifyContent, flexReversed)
				}

				if alignContent != "normal" || alignItems != "normal" {
					alignRow(rows, n, state, alignItems, alignContent)
				}

			}

			if flexDirection == "column" {
				if !flexWrapped {
					// if the container has a size restriction
					var totalHeight, maxH float32
					var fixedHeightElements int
					for i, v := range n.Children {
						vState := s[v.Properties.Id]
						if v.CStyle["min-height"] != "" {
							selfHeight -= vState.Height + vState.Margin.Top + vState.Margin.Bottom + (vState.Border.Top.Width + vState.Border.Bottom.Width)
							fixedHeightElements++
							maxH = utils.Max(maxH, vState.Height)
						} else {
							// accoutn for element min height
							totalHeight += minHeights[i] + vState.Margin.Top + vState.Margin.Bottom + (vState.Border.Top.Width + vState.Border.Bottom.Width)
							maxH = utils.Max(maxH, minHeights[i])
						}
					}

					heightDelta := selfHeight - totalHeight
					if heightDelta < 0 {
						heightDelta = -heightDelta
					}
					heightAdj := heightDelta / float32(len(n.Children)-fixedHeightElements)
					if heightAdj < 0 {
						heightAdj = -heightAdj
					}
					// We are calculating the amount a element needs to shrink because of its siblings
					for i, v := range n.Children {
						vState := s[v.Properties.Id]
						yStore := vState.Y
						vState.Height = setHeight(v, state, minHeights[i]-heightAdj)
						if i > 0 {
							sib := s[n.Children[i-1].Properties.Id]

							vState.Y = sib.Y + sib.Height + sib.Margin.Bottom + sib.Border.Top.Width + vState.Margin.Top + vState.Border.Top.Width
						}
						propagateOffsets(v, vState.X, yStore, vState.X, vState.Y, state)

						(*state)[v.Properties.Id] = vState
					}

					rows = append(rows, []int{0, len(n.Children) - 1, int(maxH)})

				} else {
					// need to redo this, with col wrap make each row the width of the longest element unless set otherwise (width, max-width) get width with min-width
					var colHeight float32
					var colIndex int
					cols := [][][]float32{}

					// Map elements to columns
					for i, v := range n.Children {
						vState := s[v.Properties.Id]
						height := vState.Height + vState.Margin.Top + vState.Margin.Bottom + (vState.Border.Top.Width + vState.Border.Bottom.Width)
						if colHeight+height > selfHeight {
							colHeight = height
							colIndex++
							width := innerSizes[i][0]
							if colIndex >= len(cols) {
								cols = append(cols, [][]float32{})
							}
							cols[colIndex] = append(cols[colIndex], []float32{float32(i), height, width})
						} else {
							colHeight += height
							width := innerSizes[i][0]
							if colIndex >= len(cols) {
								cols = append(cols, [][]float32{})
							}
							cols[colIndex] = append(cols[colIndex], []float32{float32(i), height, width})
						}
					}

					// Find the max total width of all columns
					var totalMaxWidth float32
					maxWidths := []float32{}
					for _, col := range cols {
						var maxWidth, maxHeight float32
						for _, element := range col {
							maxHeight = utils.Max(utils.Max(element[1], minHeights[int(element[0])]), maxHeight)
							maxWidth = utils.Max(element[2], maxWidth)
						}
						rows = append(rows, []int{int(col[0][0]), int(col[len(col)-1][0]), int(maxHeight)})
						totalMaxWidth += maxWidth
						maxWidths = append(maxWidths, maxWidth)
					}
					// Move the elements into the correct position
					var xOffset float32
					var index int
					for i, col := range cols {
						// Move the elements into the correct position
						yOffset := self.Y + self.Border.Top.Width + self.Padding.Top
						var marginOffset float32
						for _, element := range col {
							v := n.Children[int(element[0])]
							vState := s[v.Properties.Id]
							xStore := vState.X
							yStore := vState.Y
							vState.X = self.X + self.Padding.Left + self.Border.Left.Width + xOffset + vState.Margin.Left + vState.Border.Left.Width
							vState.Y = yOffset + vState.Margin.Top + vState.Border.Top.Width
							propagateOffsets(v, xStore, yStore, vState.X, vState.Y, state)
							if innerSizes[index][0] == maxWidths[i] {
								marginOffset = vState.Margin.Right + vState.Margin.Left + (vState.Border.Left.Width + vState.Border.Right.Width)
							}
							vState.Width = setWidth(v, state, maxWidths[i])
							yOffset += vState.Margin.Top + vState.Border.Top.Width + vState.Height + vState.Margin.Bottom + vState.Border.Bottom.Width
							(*state)[v.Properties.Id] = vState
							index++
						}
						xOffset += maxWidths[i] + marginOffset
					}

				}

				if flexReversed {
					colReverse(rows, n, state)
				}

				if justifyContent != "normal" {
					justifyCols(rows, n, state, justifyContent, flexReversed)
				}
				if alignContent != "normal" || alignItems != "normal" {
					alignCols(rows, n, state, alignItems, alignContent, innerSizes)
				}
			}
			if n.CStyle["height"] == "" && n.CStyle["min-height"] == "" {
				_, h := getInnerSize(n, state)
				self.Height = h
			}
			(*state)[n.Properties.Id] = self
		},
	}
}

func applyBlock(n *element.Node, state *map[string]element.State) {
	if len(n.Children) > 0 {
		accum := float32(0)
		inlineOffset := float32(0)
		s := *state
		lastHeight := float32(0)
		baseY := s[n.Children[0].Properties.Id].Y
		for i := 0; i < len(n.Children); i++ {
			v := n.Children[i]
			vState := s[v.Properties.Id]

			if v.CStyle["display"] != "block" {
				vState.Y += inlineOffset
				accum = (vState.Y - baseY)
				lastHeight = vState.Height
			} else if v.CStyle["position"] != "absolute" {
				vState.Y += accum
				inlineOffset += (vState.Height + (vState.Border.Top.Width + vState.Border.Bottom.Width) + vState.Margin.Top + vState.Margin.Bottom + vState.Padding.Top + vState.Padding.Bottom) + lastHeight
			}
			(*state)[v.Properties.Id] = vState
		}
	}
}

func deInline(n *element.Node, state *map[string]element.State) {
	s := *state
	// self := s[n.Properties.Id]
	baseX := float32(-1)
	baseY := float32(-1)
	for _, v := range n.Children {
		vState := s[v.Properties.Id]

		if v.CStyle["display"] == "inline" {
			if baseX < 0 && baseY < 0 {
				baseX = vState.X
				baseY = vState.Y
			} else {
				vState.X = baseX
				vState.Y = baseY
				(*state)[v.Properties.Id] = vState

			}
		} else {
			baseX = float32(-1)
			baseY = float32(-1)
		}

		if len(v.Children) > 0 {
			deInline(v, state)
		}
	}

}

func applyInline(n *element.Node, state *map[string]element.State, c *cstyle.CSS) {
	pl := inline.Init()
	for i := 0; i < len(n.Children); i++ {
		v := n.Children[i]

		if len(v.Children) > 0 {
			applyInline(v, state, c)
		}

		if pl.Selector(v) {
			pl.Handler(v, state, c)
		}
	}
}

func propagateOffsets(n *element.Node, prevx, prevy, newx, newy float32, state *map[string]element.State) {
	s := *state
	for _, v := range n.Children {
		vState := s[v.Properties.Id]
		xStore := (vState.X - prevx) + newx
		yStore := (vState.Y - prevy) + newy

		if len(v.Children) > 0 {
			propagateOffsets(v, vState.X, vState.Y, xStore, yStore, state)
		}
		vState.X = xStore
		vState.Y = yStore
		(*state)[v.Properties.Id] = vState
	}

}

func countText(n *element.Node) int {
	count := 0
	groups := []int{}
	for _, v := range n.Children {
		if v.TagName == "notaspan" {
			count += 1
		}
		if v.CStyle["display"] == "block" {
			groups = append(groups, count)
			count = 0
		}
		if len(v.Children) > 0 {
			count += countText(v)
		}
	}
	groups = append(groups, count)

	sort.Slice(groups, func(i, j int) bool {
		return groups[i] > groups[j]
	})
	return groups[0]
}

func minHeight(n *element.Node, state *map[string]element.State, prev float32) float32 {
	s := *state
	self := s[n.Properties.Id]
	if n.CStyle["min-height"] != "" {
		mw := utils.ConvertToPixels(n.CStyle["min-height"], self.EM, s[n.Parent.Properties.Id].Width)
		return utils.Max(prev, mw)
	} else {
		return prev
	}

}

func getMinHeight(n *element.Node, state *map[string]element.State) float32 {
	s := *state
	self := s[n.Properties.Id]
	selfHeight := float32(0)

	if len(n.Children) > 0 {
		for _, v := range n.Children {
			selfHeight = utils.Max(selfHeight, getNodeHeight(v, state))
		}
	} else {
		selfHeight = self.Height
	}
	if n.CStyle["min-height"] != "" {
		mh := utils.ConvertToPixels(n.CStyle["min-height"], self.EM, s[n.Parent.Properties.Id].Width)
		selfHeight = utils.Max(mh, selfHeight)
	}

	selfHeight += self.Padding.Top + self.Padding.Bottom
	return selfHeight
}

func getMinWidth(n *element.Node, state *map[string]element.State) float32 {
	s := *state
	self := s[n.Properties.Id]
	selfWidth := float32(0)

	if len(n.Children) > 0 {
		for _, v := range n.Children {
			selfWidth = utils.Max(selfWidth, getNodeWidth(v, state))
		}
	} else {
		selfWidth = self.Width
	}
	if n.CStyle["min-width"] != "" {
		mw := utils.ConvertToPixels(n.CStyle["min-width"], self.EM, s[n.Parent.Properties.Id].Width)
		selfWidth = utils.Max(mw, selfWidth)
	}

	selfWidth += self.Padding.Left + self.Padding.Right
	return selfWidth
}
func getMaxWidth(n *element.Node, state *map[string]element.State) float32 {
	s := *state
	self := s[n.Properties.Id]
	selfWidth := float32(0)

	if len(n.Children) > 0 {
		var maxRowWidth, rowWidth float32

		for _, v := range n.Children {
			rowWidth += getNodeWidth(v, state)
			if v.CStyle["display"] != "inline" {
				maxRowWidth = utils.Max(rowWidth, maxRowWidth)
				rowWidth = 0
			}
		}
		selfWidth = utils.Max(rowWidth, maxRowWidth)
	} else {
		selfWidth = self.Width
	}

	selfWidth += self.Padding.Left + self.Padding.Right
	return selfWidth
}

func getNodeWidth(n *element.Node, state *map[string]element.State) float32 {
	s := *state
	self := s[n.Properties.Id]
	w := float32(0)
	w += self.Padding.Left
	w += self.Padding.Right

	w += self.Margin.Left
	w += self.Margin.Right

	w += self.Width

	w += self.Border.Left.Width + self.Border.Right.Width

	for _, v := range n.Children {
		w = utils.Max(w, getNodeWidth(v, state))
	}

	return w
}

func setWidth(n *element.Node, state *map[string]element.State, width float32) float32 {
	s := *state
	self := s[n.Properties.Id]

	if n.CStyle["width"] != "" {
		return utils.ConvertToPixels(n.CStyle["width"], self.EM, s[n.Parent.Properties.Id].Width) + self.Padding.Left + self.Padding.Right
	}

	var maxWidth, minWidth float32
	maxWidth = 10e9
	if n.CStyle["min-width"] != "" {
		minWidth = utils.ConvertToPixels(n.CStyle["min-width"], self.EM, s[n.Parent.Properties.Id].Width)
		minWidth += self.Padding.Left + self.Padding.Right
	}
	if n.CStyle["max-width"] != "" {
		maxWidth = utils.ConvertToPixels(n.CStyle["min-width"], self.EM, s[n.Parent.Properties.Id].Width)
		maxWidth += self.Padding.Left + self.Padding.Right
	}

	return utils.Max(minWidth, utils.Min(width, maxWidth))
}

func setHeight(n *element.Node, state *map[string]element.State, height float32) float32 {
	s := *state
	self := s[n.Properties.Id]

	if n.CStyle["height"] != "" {
		return utils.ConvertToPixels(n.CStyle["height"], self.EM, s[n.Parent.Properties.Id].Height)
	}

	var maxHeight, minHeight float32
	maxHeight = 10e9
	if n.CStyle["min-height"] != "" {
		minHeight = utils.ConvertToPixels(n.CStyle["min-height"], self.EM, s[n.Parent.Properties.Id].Height)
	}
	if n.CStyle["max-height"] != "" {
		maxHeight = utils.ConvertToPixels(n.CStyle["min-height"], self.EM, s[n.Parent.Properties.Id].Height)
	}

	return utils.Max(minHeight, utils.Min(height, maxHeight))
}

func getNodeHeight(n *element.Node, state *map[string]element.State) float32 {
	s := *state
	self := s[n.Properties.Id]
	h := float32(0)
	h += self.Padding.Top
	h += self.Padding.Bottom

	h += self.Margin.Top
	h += self.Margin.Bottom

	h += self.Height

	h += self.Border.Top.Width + self.Border.Bottom.Width

	for _, v := range n.Children {
		h = utils.Max(h, getNodeHeight(v, state))
	}

	return h
}

func getInnerSize(n *element.Node, state *map[string]element.State) (float32, float32) {
	s := *state
	self := s[n.Properties.Id]

	minx := float32(10e10)
	maxw := float32(0)
	miny := float32(10e10)
	maxh := float32(0)
	for _, v := range n.Children {
		vState := s[v.Properties.Id]
		minx = utils.Min(vState.X, minx)
		miny = utils.Min(vState.Y-vState.Margin.Top, miny)
		// Don't add the top or left because the x&y values already take that into account
		hOffset := (vState.Border.Bottom.Width) + vState.Margin.Bottom
		wOffset := (vState.Border.Right.Width) + vState.Margin.Right
		maxw = utils.Max(vState.X+vState.Width+wOffset, maxw)
		maxh = utils.Max(vState.Y+vState.Height+hOffset, maxh)
	}
	w := maxw - minx
	h := maxh - miny

	w += self.Padding.Left + self.Padding.Right
	h += self.Padding.Top + self.Padding.Bottom
	if n.CStyle["width"] != "" {
		w = self.Width
	}
	if n.CStyle["height"] != "" {
		h = self.Height
	}

	return w, h
}

func add2d(arr [][]float32, index int) float32 {
	var sum float32
	if len(arr) == 0 {
		return sum
	}

	for i := 0; i < len(arr); i++ {
		if len(arr[i]) <= index {
			return sum
		}
		sum += arr[i][index]
	}

	return sum
}

func colReverse(cols [][]int, n *element.Node, state *map[string]element.State) {
	s := *state
	for _, col := range cols {
		tempNodes := []*element.Node{}
		tempStates := []element.State{}

		for i := col[1]; i >= col[0]; i-- {
			tempNodes = append(tempNodes, n.Children[i])
			tempStates = append(tempStates, s[n.Children[i].Properties.Id])
		}

		for i := 0; i < len(tempStates); i++ {
			e := col[0] + i
			vState := s[n.Children[e].Properties.Id]
			propagateOffsets(n.Children[e], vState.X, vState.Y, tempStates[i].X, tempStates[i].Y, state)
			vState.Y = tempStates[i].Y
			(*state)[n.Children[e].Properties.Id] = vState
		}
		for i := 0; i < len(tempStates); i++ {
			e := col[0] + i
			n.Children[e] = tempNodes[i]
		}
	}
}

func rowReverse(rows [][]int, n *element.Node, state *map[string]element.State) {
	s := *state
	for _, row := range rows {
		tempNodes := []*element.Node{}
		tempStates := []element.State{}

		for i := row[1] - 1; i >= row[0]; i-- {
			tempNodes = append(tempNodes, n.Children[i])
			tempStates = append(tempStates, s[n.Children[i].Properties.Id])
		}

		for i := 0; i < len(tempStates); i++ {
			e := row[0] + i
			vState := s[n.Children[e].Properties.Id]
			propagateOffsets(n.Children[e], vState.X, vState.Y, tempStates[i].X, tempStates[i].Y, state)
			vState.X = tempStates[i].X
			(*state)[n.Children[e].Properties.Id] = vState
		}
		for i := 0; i < len(tempStates); i++ {
			e := row[0] + i
			n.Children[e] = tempNodes[i]
		}

		for i := row[1] - 1; i >= row[0]; i-- {
			vState := s[n.Children[i].Properties.Id]
			var xChng float32
			if i < row[1]-1 {
				sib := s[n.Children[i+1].Properties.Id]
				xChng = sib.X - (sib.Border.Left.Width + sib.Margin.Left + vState.Margin.Right + vState.Border.Right.Width + vState.Width)
			} else {
				parent := s[n.Properties.Id]
				xChng = ((((parent.X + parent.Width) - parent.Padding.Right) - vState.Width) - vState.Margin.Right) - (vState.Border.Right.Width)

			}
			propagateOffsets(n.Children[i], vState.X, vState.Y, xChng, vState.Y, state)
			vState.X = xChng
			(*state)[n.Children[i].Properties.Id] = vState
		}
	}
}

func justifyRow(rows [][]int, n *element.Node, state *map[string]element.State, justify string, reversed bool) {
	s := *state
	for _, row := range rows {

		if (justify == "flex-end" || justify == "end" || justify == "right") && !reversed {
			for i := row[1] - 1; i >= row[0]; i-- {
				vState := s[n.Children[i].Properties.Id]
				var xChng float32
				if i < row[1]-1 {
					sib := s[n.Children[i+1].Properties.Id]
					xChng = sib.X - (sib.Border.Left.Width + sib.Margin.Left + vState.Margin.Right + vState.Border.Left.Width + vState.Width)
				} else {
					parent := s[n.Properties.Id]
					xChng = ((((parent.X + parent.Width) - parent.Padding.Right) - vState.Width) - vState.Margin.Right) - (vState.Border.Right.Width)

				}
				propagateOffsets(n.Children[i], vState.X, vState.Y, xChng, vState.Y, state)
				vState.X = xChng
				(*state)[n.Children[i].Properties.Id] = vState
			}
		} else if (justify == "flex-end" || justify == "start" || justify == "left" || justify == "normal") && reversed {
			for i := row[0]; i < row[1]; i++ {
				vState := s[n.Children[i].Properties.Id]
				var xChng float32
				if i > row[0] {
					sib := s[n.Children[i-1].Properties.Id]
					xChng = sib.X + sib.Width + (sib.Border.Left.Width + sib.Border.Right.Width) + sib.Margin.Right + vState.Margin.Left + vState.Border.Left.Width
				} else {
					parent := s[n.Properties.Id]
					xChng = parent.X + parent.Padding.Right + vState.Margin.Left + vState.Border.Left.Width + parent.Border.Right.Width

				}
				propagateOffsets(n.Children[i], vState.X, vState.Y, xChng, vState.Y, state)
				vState.X = xChng
				(*state)[n.Children[i].Properties.Id] = vState
			}
		} else if justify == "center" {
			// get width of row then center (by getting last x + w + mr + b)
			f := s[n.Children[row[0]].Properties.Id]
			l := s[n.Children[row[1]-1].Properties.Id]
			parent := s[n.Properties.Id]
			po := parent.X + parent.Border.Left.Width
			// ?
			offset := (parent.Width - ((f.X - po) + (l.X - po) + l.Width + f.Border.Left.Width + l.Border.Right.Width)) / 2

			for i := row[0]; i < row[1]; i++ {
				vState := s[n.Children[i].Properties.Id]

				if !reversed {
					propagateOffsets(n.Children[i], vState.X, vState.Y, vState.X+offset, vState.Y, state)
					vState.X += offset
				} else {
					propagateOffsets(n.Children[i], vState.X, vState.Y, vState.X-offset, vState.Y, state)
					vState.X -= offset
				}
				(*state)[n.Children[i].Properties.Id] = vState
			}

		} else if justify == "space-between" {
			// get width of row then center (by getting last x + w + mr + b)
			f := s[n.Children[row[0]].Properties.Id]
			l := s[n.Children[row[1]-1].Properties.Id]
			parent := s[n.Properties.Id]
			po := parent.Border.Left.Width + parent.Width
			po -= parent.Padding.Left + parent.Padding.Right

			// make po repersent the total space between elements
			for i := row[0]; i < row[1]; i++ {
				vState := s[n.Children[i].Properties.Id]
				po -= vState.Width + vState.Margin.Left + vState.Margin.Right + (vState.Border.Left.Width + vState.Border.Right.Width)
			}

			po /= float32(((row[1]) - row[0]) - 1)

			if (row[1]-1)-row[0] > 0 {
				for i := row[0]; i < row[1]; i++ {
					vState := s[n.Children[i].Properties.Id]
					var offset float32
					if i == row[0] {
						offset = parent.X + parent.Padding.Left + f.Margin.Left + f.Border.Left.Width
					} else if i == row[1]-1 {
						offset = (parent.X + parent.Width) - (l.Margin.Right + l.Border.Left.Width + l.Width + parent.Padding.Right)
					} else {
						if !reversed {
							offset = vState.X + (po * float32(i-row[0]))
						} else {
							offset = vState.X - (po * float32(((row[1]-1)-row[0])-(i-row[0])))
						}

					}

					propagateOffsets(n.Children[i], vState.X, vState.Y, offset, vState.Y, state)
					vState.X = offset
					(*state)[n.Children[i].Properties.Id] = vState
				}
			}

		} else if justify == "space-evenly" {
			// get width of row then center (by getting last x + w + mr + b)
			parent := s[n.Properties.Id]
			po := parent.Border.Left.Width + parent.Width
			po -= parent.Padding.Left + parent.Padding.Right

			// make po repersent the total space between elements
			for i := row[0]; i < row[1]; i++ {
				vState := s[n.Children[i].Properties.Id]
				po -= vState.Width + vState.Margin.Left + vState.Margin.Right + (vState.Border.Left.Width + vState.Border.Right.Width)
			}

			po /= float32(((row[1]) - row[0]) + 1)

			// get width of row then center (by getting last x + w + mr + b)

			for i := row[0]; i < row[1]; i++ {
				vState := s[n.Children[i].Properties.Id]

				if !reversed {
					offset := po * (float32(i-row[0]) + 1)
					propagateOffsets(n.Children[i], vState.X, vState.Y, vState.X+offset, vState.Y, state)
					vState.X += offset
				} else {
					offset := po * float32(((row[1]-1)-row[0])-((i-row[0])-1))

					propagateOffsets(n.Children[i], vState.X, vState.Y, vState.X-offset, vState.Y, state)
					vState.X -= offset
				}
				(*state)[n.Children[i].Properties.Id] = vState
			}

		} else if justify == "space-around" {
			// get width of row then center (by getting last x + w + mr + b)
			parent := s[n.Properties.Id]
			po := parent.Border.Left.Width + parent.Width
			po -= parent.Padding.Left + parent.Padding.Right

			// make po repersent the total space between elements
			for i := row[0]; i < row[1]; i++ {
				vState := s[n.Children[i].Properties.Id]
				po -= vState.Width + vState.Margin.Left + vState.Margin.Right + (vState.Border.Left.Width + vState.Border.Right.Width)
			}

			po /= float32(((row[1]) - row[0]))

			// get width of row then center (by getting last x + w + mr + b)

			for i := row[0]; i < row[1]; i++ {
				vState := s[n.Children[i].Properties.Id]

				if !reversed {
					m := (float32(i-row[0]) + 1)
					if i-row[0] == 0 {
						m = 0.5
					} else {
						m -= 0.5
					}
					offset := po * m
					propagateOffsets(n.Children[i], vState.X, vState.Y, vState.X+offset, vState.Y, state)
					vState.X += offset
				} else {
					m := float32(((row[1] - 1) - row[0]) - ((i - row[0]) - 1))
					m -= 0.5
					offset := po * m

					propagateOffsets(n.Children[i], vState.X, vState.Y, vState.X-offset, vState.Y, state)
					vState.X -= offset
				}
				(*state)[n.Children[i].Properties.Id] = vState
			}

		}

	}
}

func alignRow(rows [][]int, n *element.Node, state *map[string]element.State, align, content string) {
	// !ISSUE: Baseline isn't properly impleamented
	s := *state
	self := s[n.Properties.Id]

	maxes := []float32{}
	var maxesTotal float32
	for _, row := range rows {
		var maxH float32
		for i := row[0]; i < row[1]; i++ {
			v := n.Children[i]
			vState := s[v.Properties.Id]
			_, h := getInnerSize(v, state)
			h = minHeight(v, state, h)
			h = setHeight(v, state, h)
			vState.Height = h
			h += vState.Margin.Top + vState.Margin.Bottom + (vState.Border.Top.Width + vState.Border.Bottom.Width)
			maxH = utils.Max(maxH, h)
			(*state)[v.Properties.Id] = vState
		}
		maxes = append(maxes, maxH)
		maxesTotal += maxH
	}

	os := ((self.Height - (self.Padding.Top + self.Padding.Bottom + (self.Border.Top.Width + self.Border.Bottom.Width))) - maxesTotal) / float32(len(rows))
	if os < 0 || content != "normal" {
		os = 0
	}

	var contentOffset float32

	if content == "center" {
		contentOffset = ((self.Height - (self.Padding.Top + self.Padding.Bottom + (self.Border.Top.Width + self.Border.Bottom.Width))) - maxesTotal) / 2
	} else if content == "end" || content == "flex-end" {
		contentOffset = ((self.Height - (self.Padding.Top + self.Padding.Bottom + (self.Border.Top.Width + self.Border.Bottom.Width))) - maxesTotal)
	} else if content == "start" || content == "flex-start" || content == "baseline" {
		// This is redundent but it helps keep track
		contentOffset = 0
	} else if content == "space-between" {
		os = ((self.Height - (self.Padding.Top + self.Padding.Bottom + (self.Border.Top.Width + self.Border.Bottom.Width))) - maxesTotal) / float32(len(rows)-1)
	} else if content == "space-around" {
		os = ((self.Height - (self.Padding.Top + self.Padding.Bottom + (self.Border.Top.Width + self.Border.Bottom.Width))) - maxesTotal) / float32(len(rows))
		contentOffset = os / 2
	} else if content == "space-evenly" {
		os = ((self.Height - (self.Padding.Top + self.Padding.Bottom + (self.Border.Top.Width + self.Border.Bottom.Width))) - maxesTotal) / float32(len(rows)+1)
		contentOffset = os
	}

	for c, row := range rows {
		maxH := maxes[c]
		var sum float32
		for i := 0; i < c; i++ {
			sum += maxes[i]
		}
		if align == "start" || align == "flex-start" || align == "self-start" || align == "normal" {
			for i := row[0]; i < row[1]; i++ {
				vState := s[n.Children[i].Properties.Id]

				offset := sum + self.Y + self.Padding.Top + vState.Margin.Top + contentOffset

				if n.CStyle["height"] != "" || n.CStyle["min-height"] != "" {
					offset += ((os) * float32(c))
				}

				propagateOffsets(n.Children[i], vState.X, vState.Y, vState.X, offset, state)
				vState.Y = offset
				(*state)[n.Children[i].Properties.Id] = vState
			}
		} else if align == "center" {
			for i := row[0]; i < row[1]; i++ {
				vState := s[n.Children[i].Properties.Id]

				offset := sum + self.Y + self.Padding.Top + vState.Margin.Top + contentOffset

				if n.CStyle["height"] != "" || n.CStyle["min-height"] != "" {
					offset += (os * float32(c+1)) - (os / 2)
				}

				if vState.Height+vState.Margin.Top+vState.Margin.Bottom+(vState.Border.Top.Width+vState.Border.Bottom.Width) < maxH {
					offset += (maxH - (vState.Height + vState.Margin.Top + vState.Margin.Bottom + (vState.Border.Top.Width + vState.Border.Bottom.Width))) / 2
				}
				propagateOffsets(n.Children[i], vState.X, vState.Y, vState.X, offset, state)
				vState.Y = offset
				(*state)[n.Children[i].Properties.Id] = vState
			}
		} else if align == "end" || align == "flex-end" || align == "self-end" {
			for i := row[0]; i < row[1]; i++ {
				vState := s[n.Children[i].Properties.Id]

				offset := sum + self.Y + self.Padding.Top + vState.Margin.Top + contentOffset

				if n.CStyle["height"] != "" || n.CStyle["min-height"] != "" {
					offset += os * float32(c+1)
				}

				if vState.Height+vState.Margin.Top+vState.Margin.Bottom+(vState.Border.Top.Width+vState.Border.Bottom.Width) < maxH {
					offset += (maxH - (vState.Height + vState.Margin.Top + vState.Margin.Bottom + (vState.Border.Top.Width + vState.Border.Bottom.Width)))
				}
				propagateOffsets(n.Children[i], vState.X, vState.Y, vState.X, offset, state)
				vState.Y = offset
				(*state)[n.Children[i].Properties.Id] = vState

			}
		} else if align == "stretch" {
			for i := row[0]; i < row[1]; i++ {
				vState := s[n.Children[i].Properties.Id]

				offset := sum + self.Y + self.Padding.Top + vState.Margin.Top

				if n.CStyle["height"] != "" || n.CStyle["min-height"] != "" {
					offset += ((os) * float32(c))
				}

				propagateOffsets(n.Children[i], vState.X, vState.Y, vState.X, offset, state)
				vState.Y = offset
				vState.Height = maxH - (vState.Margin.Top + vState.Margin.Bottom + (vState.Border.Top.Width + vState.Border.Bottom.Width))
				(*state)[n.Children[i].Properties.Id] = vState

			}
		}
	}
}

func justifyCols(cols [][]int, n *element.Node, state *map[string]element.State, justify string, reversed bool) {
	s := *state
	self := s[n.Properties.Id]

	selfHeight := (self.Height) - (self.Padding.Top + self.Padding.Bottom)
	for _, col := range cols {
		yCollect := self.Y + self.Padding.Top
		var colHeight float32
		for i := col[0]; i <= col[1]; i++ {
			v := n.Children[i]
			vState := s[v.Properties.Id]
			colHeight += vState.Height + vState.Margin.Top + vState.Margin.Bottom + (vState.Border.Top.Width + vState.Border.Bottom.Width)
		}

		if justify == "center" {
			offset := ((selfHeight - colHeight) / 2)
			yCollect += offset
			for i := col[0]; i <= col[1]; i++ {
				v := n.Children[i]
				vState := s[v.Properties.Id]
				yStore := vState.Y
				vState.Y = yCollect + vState.Margin.Top
				yCollect += vState.Height + vState.Margin.Bottom + vState.Margin.Top + vState.Border.Top.Width + vState.Border.Bottom.Width
				propagateOffsets(n.Children[i], vState.X, yStore, vState.X, vState.Y, state)
				(*state)[v.Properties.Id] = vState
			}
		}

		if justify == "end" || justify == "flex-end" {
			offset := (selfHeight - colHeight)
			yCollect += offset
			for i := col[0]; i <= col[1]; i++ {
				v := n.Children[i]
				vState := s[v.Properties.Id]
				yStore := vState.Y
				vState.Y = yCollect + vState.Border.Top.Width + vState.Margin.Top
				yCollect += vState.Height + vState.Margin.Bottom + vState.Border.Top.Width + vState.Margin.Top + vState.Border.Bottom.Width
				propagateOffsets(n.Children[i], vState.X, yStore, vState.X, vState.Y, state)
				(*state)[v.Properties.Id] = vState
			}
		}

		if justify == "space-evenly" {
			offset := (selfHeight - colHeight) / (float32(col[1]-col[0]) + 2)
			for i := col[0]; i <= col[1]; i++ {
				v := n.Children[i]
				vState := s[v.Properties.Id]
				yStore := vState.Y
				vState.Y = yCollect + vState.Border.Top.Width + vState.Margin.Top + offset
				yCollect += vState.Height + vState.Margin.Bottom + vState.Border.Top.Width + vState.Margin.Top + vState.Border.Bottom.Width + offset
				propagateOffsets(n.Children[i], vState.X, yStore, vState.X, vState.Y, state)
				(*state)[v.Properties.Id] = vState
			}
		}

		if justify == "space-between" {
			offset := (selfHeight - colHeight) / (float32(col[1] - col[0]))
			for i := col[0]; i <= col[1]; i++ {
				v := n.Children[i]
				vState := s[v.Properties.Id]
				yStore := vState.Y
				vState.Y = yCollect + vState.Border.Top.Width + vState.Margin.Top
				if col[1]-col[0] != 0 {
					vState.Y += offset * float32(i-col[0])
				} else if reversed {
					vState.Y += selfHeight - (vState.Height + vState.Margin.Bottom + vState.Border.Top.Width + vState.Margin.Top + vState.Border.Bottom.Width)
				}
				yCollect += vState.Height + vState.Margin.Bottom + vState.Border.Top.Width + vState.Margin.Top + vState.Border.Bottom.Width
				propagateOffsets(n.Children[i], vState.X, yStore, vState.X, vState.Y, state)
				(*state)[v.Properties.Id] = vState
			}
		}
		if justify == "space-around" {
			offset := (selfHeight - colHeight) / (float32(col[1]-col[0]) + 1)
			for i := col[0]; i <= col[1]; i++ {
				v := n.Children[i]
				vState := s[v.Properties.Id]
				yStore := vState.Y
				vState.Y = yCollect + vState.Border.Top.Width + vState.Margin.Top
				if col[1]-col[0] == 0 {
					vState.Y += offset / 2
				} else {
					vState.Y += (offset * float32(i-col[0])) + (offset / 2)
				}
				yCollect += vState.Height + vState.Margin.Bottom + vState.Border.Top.Width + vState.Margin.Top + vState.Border.Bottom.Width
				propagateOffsets(n.Children[i], vState.X, yStore, vState.X, vState.Y, state)
				(*state)[v.Properties.Id] = vState
			}
		}
	}
}

func alignCols(cols [][]int, n *element.Node, state *map[string]element.State, align, content string, minWidths [][]float32) {
	s := *state
	self := s[n.Properties.Id]

	if (align == "stretch" && content == "stretch") || (align == "stretch" && content == "normal") || (align == "normal" && content == "stretch") {
		return
	}

	selfWidth := (self.Width - self.Padding.Left) - self.Padding.Right

	var rowWidth float32
	colWidths := []float32{}
	for _, col := range cols {
		var maxWidth float32
		for i := col[0]; i <= col[1]; i++ {
			v := n.Children[i]
			vState := s[v.Properties.Id]
			vState.Width = setWidth(v, state, minWidths[i][0])
			maxWidth = utils.Max(maxWidth, vState.Width+vState.Margin.Left+vState.Margin.Right+(vState.Border.Left.Width+vState.Border.Right.Width))
			(*state)[v.Properties.Id] = vState
		}
		colWidths = append(colWidths, maxWidth)
		rowWidth += maxWidth
	}
	var xOffset float32
	for c, col := range cols {

		if align != "normal" {
			// var offset float32
			for i := col[0]; i <= col[1]; i++ {
				v := n.Children[i]
				vState := s[v.Properties.Id]
				xStore := vState.X
				vState.X = self.X + self.Padding.Left + self.Border.Left.Width + xOffset + vState.Margin.Left + vState.Border.Right.Width
				if align == "center" {
					vState.X += ((colWidths[c] - (vState.Width + vState.Margin.Left + vState.Margin.Right + (vState.Border.Left.Width + vState.Border.Right.Width))) / 2)
				} else if align == "end" || align == "flex-end" || align == "self-end" {
					vState.X += (colWidths[c] - (vState.Width + vState.Margin.Left + vState.Margin.Right + (vState.Border.Left.Width + vState.Border.Right.Width)))
				} else if align == "stretch" {
					vState.Width = (colWidths[c] - (vState.Margin.Left + vState.Margin.Right + (vState.Border.Left.Width + vState.Border.Right.Width)))
				}
				propagateOffsets(n.Children[i], xStore, vState.Y, vState.X, vState.Y, state)
				(*state)[v.Properties.Id] = vState
			}
			xOffset += colWidths[c]
		} else {
			// if align is set to normal then make the elements the size of the column
			// the size of the column is
			for i := col[0]; i <= col[1]; i++ {
				v := n.Children[i]
				vState := s[v.Properties.Id]
				vState.Width = setWidth(v, state, colWidths[c]-(vState.Margin.Left+vState.Margin.Right+(vState.Border.Left.Width+vState.Border.Right.Width)))
				(*state)[v.Properties.Id] = vState
			}
		}
	}

	var offset float32
	if selfWidth < rowWidth {
		selfWidth = rowWidth
	}
	if content == "center" {
		offset = ((selfWidth - rowWidth) / 2)
	}
	if content == "end" || content == "flex-end" {
		offset = (selfWidth - rowWidth)
	}
	if content == "space-evenly" {
		offset = (selfWidth - rowWidth) / (float32(len(cols) + 1))
	}
	if content == "space-between" {
		offset = (selfWidth - rowWidth) / (float32(len(cols) - 1))
	}
	if content == "space-around" {
		offset = (selfWidth - rowWidth) / (float32(len(cols)))
	}
	if content == "normal" || content == "stretch" {
		if align == "start" || align == "end" || content == "flex-end" {
			offset = (selfWidth - rowWidth) / (float32(len(cols)))
		}
		if align == "center" {
			offset = ((selfWidth - rowWidth) / (float32(len(cols))))
		}
	}
	for c, col := range cols {

		for i := col[0]; i <= col[1]; i++ {
			v := n.Children[i]
			vState := s[v.Properties.Id]
			xStore := vState.X
			if content == "center" || content == "end" || content == "flex-end" {
				vState.X += offset
			} else if content == "space-evenly" {
				vState.X += offset * float32(c+1)
			} else if content == "space-between" {
				vState.X += offset * float32(c)
			} else if content == "space-around" {
				if c == 0 {
					vState.X += offset / 2
				} else {
					vState.X += (offset * float32(c)) + (offset / 2)
				}
			} else if content == "normal" || content == "stretch" {
				if align == "start" {
					vState.X += offset * float32(c)
				}
				if align == "end" || content == "flex-end" {
					vState.X += offset * float32(c+1)
				}
				if align == "center" {
					if c == 0 {
						vState.X += offset / 2
					} else {
						vState.X += (offset * float32(c)) + (offset / 2)
					}
				}
			}
			propagateOffsets(n.Children[i], xStore, vState.Y, vState.X, vState.Y, state)
			(*state)[v.Properties.Id] = vState
		}
	}
}
