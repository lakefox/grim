package cstyle

import (
	"bytes"
	"golang.org/x/image/draw"
	"grim/canvas"
	"grim/color"
	"grim/element"
	"grim/gg"
	"grim/utils"
	"image"
	ic "image/color"
	_ "image/gif"  // Enable GIF support
	_ "image/jpeg" // Enable JPEG support
	_ "image/png"  // Enable PNG support
	"math"
	"path/filepath"
	"strconv"
	"strings"
)

// type Background struct {
// 	Color      ic.RGBA
// 	Image      string
// 	PositionX  string
// 	PositionY  string
// 	Size       string
// 	Repeat     string
// 	Origin     string
// 	Attachment string
// }

var backgroundProps = []string{
	"background-image",
	"background-position-x",
	"background-position-y",
	"background-size",
	"background-repeat",
	"background-attachment",
	"background-origin",
	// "background-color",
}

func ParseBackground(style map[string]string) []element.Background {
	splitProps := map[string][]string{}

	amount := 0
	for _, v := range backgroundProps {
		s := utils.SplitByComma(style[v])
		f := []string{}
		for _, b := range s {
			if strings.TrimSpace(b) != "" {
				f = append(f, b)
			}
		}
		c := len(f)
		if amount < c {
			amount = c
		}
		splitProps[v] = f
	}
	if amount == 0 {
		amount = 1
	}
	bgs := []element.Background{}
	for i := range amount {
		bg := element.Background{}
		if style["background-color"] != "" {
			bg.Color, _ = color.ParseRGBA(style["background-color"])
		}

		if len(splitProps["background-image"]) != 0 {
			bg.Image = splitProps["background-image"][i]
		}

		if len(splitProps["background-position-x"])-1 >= i {
			bg.PositionX = splitProps["background-position-x"][i]
		} else {
			bg.PositionX = "0px"
		}

		if len(splitProps["background-position-y"])-1 >= i {
			bg.PositionY = splitProps["background-position-y"][i]
		} else {
			bg.PositionY = "0px"
		}

		if len(splitProps["background-size"])-1 >= i {
			bg.Size = splitProps["background-size"][i]
		} else {
			bg.Size = "auto auto"
		}

		if len(splitProps["background-repeat"])-1 >= i {
			bg.Repeat = splitProps["background-repeat"][i]
		} else {
			bg.Repeat = "repeat"
		}

		if len(splitProps["background-origin"])-1 >= i {
			bg.Origin = splitProps["background-origin"][i]
		} else {
			bg.Origin = "padding-box"
		}

		if len(splitProps["background-attachment"])-1 >= i {
			bg.Attachment = splitProps["background-attachment"][i]
		} else {
			bg.Attachment = "scroll"
		}
		bgs = append(bgs, bg)
	}
	return bgs
}

func BackgroundKey(self element.State) string {
	key := strconv.Itoa(int(self.Width)) + strconv.Itoa(int(self.Height)) + strconv.Itoa(len(self.Background))

	for _, v := range self.Background {
		key += v.Image
		key += v.PositionX
		key += v.PositionY
		key += v.Size
		key += v.Repeat
		key += v.Origin
		key += v.Attachment
		key += strconv.Itoa(int(v.Color.R)) + strconv.Itoa(int(v.Color.G)) + strconv.Itoa(int(v.Color.B)) + strconv.Itoa(int(v.Color.A))
	}

	return key
}

// !NOTE: background-clip and background-blend-mode are not supported
func GenerateBackground(c CSS, self element.State) image.Image {
	wbw := int(self.Width + self.Border.Left.Width + self.Border.Right.Width)
	hbw := int(self.Height + self.Border.Top.Width + self.Border.Bottom.Width)

	can := canvas.NewCanvas(wbw, hbw)
	for _, bg := range self.Background {
		if bg.Color.A > 0 {
			// Draw the solid background color if it is not completely tranparent
			can.BeginPath()
			can.SetFillStyle(bg.Color.R, bg.Color.G, bg.Color.B, bg.Color.A)
			can.SetLineWidth(10)
			can.RoundedRect(0, 0, float64(wbw), float64(hbw),
				[]float64{float64(self.Border.Radius.TopLeft), float64(self.Border.Radius.TopRight), float64(self.Border.Radius.BottomRight), float64(self.Border.Radius.BottomLeft)})
			can.Fill()
			can.ClosePath()
		}
		sHeight := self.Height + self.Border.Top.Width + self.Border.Bottom.Width
		sWidth := self.Width + self.Border.Left.Width + self.Border.Right.Width

		// Load background image url
		if bg.Image != "" && bg.Image != "none" {
			if bg.Image[0:4] == "url(" {

				filePath := filepath.Join(c.Path, bg.Image[5:len(bg.Image)-2])
				file, _ := c.Adapter.FileSystem.ReadFile(filePath)
				img, _, _ := image.Decode(bytes.NewReader(file))

				b := img.Bounds()
				width := int(b.Dx())
				height := int(b.Dy())

				if bg.Size != "" {

					imageRatio := float32(width / height)
					selfRatio := sWidth / sHeight

					if bg.Size == "cover" {
						if imageRatio > selfRatio {
							height = int(sHeight)
							width = int(sHeight * imageRatio)
						} else {
							width = int(sWidth)
							height = int(sWidth / imageRatio)
						}
					} else if bg.Size == "contain" {
						if imageRatio > selfRatio {
							// Image is wider proportionally - constrain by width
							width = int(sWidth)
							height = int(float32(width) / imageRatio)
						} else {
							// Image is taller proportionally - constrain by height
							height = int(sHeight)
							width = int(float32(height) * imageRatio)
						}
					} else {
						parts := strings.Split(bg.Size, " ")
						if len(parts) == 2 {
							var w, h float32
							w = utils.ConvertToPixels(parts[0], self.EM, sWidth)

							wAuto := parts[0] == "auto"

							h = utils.ConvertToPixels(parts[1], self.EM, sHeight)
							hAuto := parts[1] == "auto"

							if wAuto && hAuto {
								if selfRatio > imageRatio {
									h = float32(sHeight)
									w = h * imageRatio
								} else {
									w = float32(sWidth)
									h = w / imageRatio
								}
							} else if wAuto {
								w = h * imageRatio
							} else if hAuto {
								h = w / imageRatio
							}

							width = int(w)
							height = int(h)
						} else if len(parts) == 1 {
							var s float32
							if parts[0] == "auto" {

								if selfRatio > imageRatio {
									s = float32(sHeight) / float32(height)
								} else {
									s = float32(sWidth) / float32(width)
								}

								width = int(float32(width) * s)
								height = int(float32(height) * s)
							} else {
								s = utils.ConvertToPixels(parts[0], self.EM, sWidth)
								height = int(s * float32(height/width))
								width = int(s)
							}
						}
					}
				}
				x := 0
				y := 0

				if bg.Origin != "" {
					if bg.Origin == "padding-box" {
						x += int(self.Border.Left.Width)
						sW := int(sWidth - self.Border.Right.Width)
						sWidth = float32(sW)
						if width > sW {
							width = sW
						} else {
							width += x
						}

						y += int(self.Border.Top.Width)
						sH := int(sHeight - self.Border.Bottom.Width)
						sHeight = float32(sH)
						if height > sH {
							height = sH
						} else {
							height += y
						}
					} else if bg.Origin == "content-box" {
						x += int(self.Border.Left.Width + self.Padding.Left)
						sW := int(sWidth - (self.Border.Left.Width + self.Padding.Left + self.Border.Right.Width + self.Padding.Left))
						sWidth = float32(sW)
						if width > sW {
							width = sW
						}

						y += int(self.Border.Top.Width + self.Padding.Top)
						sH := int(sHeight - (self.Border.Top.Width + self.Padding.Top + self.Border.Bottom.Width + self.Padding.Top))
						sHeight = float32(sH)
						if height > sH {
							height = sH
						}
					}
				}

				if bg.PositionX != "" {
					parts := strings.Split(bg.PositionX, " ")

					if len(parts) == 2 {
						if parts[0] == "left" {
							d := int(utils.ConvertToPixels(parts[1], self.EM, self.Width))
							x += d
						} else if parts[0] == "right" {
							d := int(utils.ConvertToPixels(parts[1], self.EM, self.Width))
							x += int(sWidth) - (width + d)

						}
					} else if len(parts) == 1 {
						if parts[0] == "right" {
							x += int(sWidth) - (width)

						} else if parts[0] == "center" {
							x += int(sWidth/2) - (width / 2)
						} else {
							d := int(utils.ConvertToPixels(parts[0], self.EM, self.Width))
							x += d
						}
					}
				}
				if bg.PositionY != "" {
					parts := strings.Split(bg.PositionY, " ")

					if len(parts) == 2 {
						if parts[0] == "top" {
							d := int(utils.ConvertToPixels(parts[1], self.EM, self.Height))
							y += d
						} else if parts[0] == "bottom" {
							d := int(utils.ConvertToPixels(parts[1], self.EM, self.Height))
							y += int(sHeight) - (height + d)
						}
					} else if len(parts) == 1 {
						if parts[0] == "bottom" {
							y += int(sHeight) - (height)

						} else if parts[0] == "center" {
							y += int(sHeight/2) - (height / 2)
						} else {
							d := int(utils.ConvertToPixels(parts[0], self.EM, self.Height))
							y += d
						}
					}
				}

				resized := image.NewRGBA(image.Rect(0, 0, width, height))
				draw.CatmullRom.Scale(resized, resized.Bounds(), img, img.Bounds(), draw.Over, &draw.Options{})

				if bg.Repeat != "" {
					parts := strings.Split(bg.Repeat, " ")
					canvasWidth := sWidth
					canvasHeight := sHeight

					// Set default values for horizontal and vertical repeat
					repeatX := "no-repeat"
					repeatY := "no-repeat"

					// Handle single-value repeat
					if len(parts) == 1 {
						if parts[0] == "repeat-x" {
							repeatX = "repeat"
						} else if parts[0] == "repeat-y" {
							repeatY = "repeat"
						} else if parts[0] == "repeat" || parts[0] == "space" || parts[0] == "round" {
							repeatX = parts[0]
							repeatY = parts[0]
						}
					} else if len(parts) == 2 {
						repeatX = parts[0]
						repeatY = parts[1]
					}

					// Handle repeat-x and repeat-y as x/y directional repeats
					if repeatX == "repeat" && repeatY == "repeat" {
						// Calculate proper starting point to maintain pattern alignment
						tilesBackX := int(math.Ceil(float64(x) / float64(width)))
						tilesBackY := int(math.Ceil(float64(y) / float64(height)))

						startX := float64(x) - float64(tilesBackX*width)
						startY := float64(y) - float64(tilesBackY*height)

						// Draw covering the entire canvas
						for currentY := startY; currentY < float64(canvasHeight); currentY += float64(height) {
							for currentX := startX; currentX < float64(canvasWidth); currentX += float64(width) {
								can.DrawImage(resized, currentX, currentY)
							}
						}
					} else if repeatX == "repeat" && repeatY == "no-repeat" {
						// Repeat only horizontally at the specified y
						tilesBackX := int(math.Ceil(float64(x) / float64(width)))
						startX := float64(x) - float64(tilesBackX*width)
						startY := float64(y)

						for currentX := startX; currentX < float64(canvasWidth); currentX += float64(width) {
							can.DrawImage(resized, currentX, startY)
						}
					} else if repeatX == "no-repeat" && repeatY == "repeat" {
						// Repeat only vertically at the specified x
						tilesBackY := int(math.Ceil(float64(y) / float64(height)))
						startX := float64(x)
						startY := float64(y) - float64(tilesBackY*height)
						for currentY := startY; currentY < float64(canvasHeight); currentY += float64(height) {
							can.DrawImage(resized, startX, currentY)
						}
					} else if repeatX == "space" || repeatY == "space" {
						// Handle space - tiles with equal spacing, no clipping
						if repeatX == "space" && repeatY != "space" {
							// Only horizontal spacing
							availableWidth := float64(canvasWidth)
							numTiles := int(math.Floor(availableWidth / float64(width)))

							if numTiles > 1 {
								spacing := (availableWidth - float64(numTiles*width)) / float64(numTiles-1)
								for i := 0; i < numTiles; i++ {
									posX := float64(i) * (float64(width) + spacing)
									can.DrawImage(resized, posX, float64(y))
								}
							} else {
								// Only one tile - center it
								posX := (availableWidth - float64(width)) / 2
								can.DrawImage(resized, posX, float64(y))
							}
						} else if repeatX != "space" && repeatY == "space" {
							// Only vertical spacing
							availableHeight := float64(canvasHeight)
							numTiles := int(math.Floor(availableHeight / float64(height)))

							if numTiles > 1 {
								spacing := (availableHeight - float64(numTiles*height)) / float64(numTiles-1)
								for i := 0; i < numTiles; i++ {
									posY := float64(i) * (float64(height) + spacing)
									can.DrawImage(resized, float64(x), posY)
								}
							} else {
								// Only one tile - center it
								posY := (availableHeight - float64(height)) / 2
								can.DrawImage(resized, float64(x), posY)
							}
						} else {
							// Both directions use space
							availableWidth := float64(canvasWidth)
							availableHeight := float64(canvasHeight)
							numTilesX := int(math.Floor(availableWidth / float64(width)))
							numTilesY := int(math.Floor(availableHeight / float64(height)))

							spacingX := 0.0
							spacingY := 0.0

							if numTilesX > 1 {
								spacingX = (availableWidth - float64(numTilesX*width)) / float64(numTilesX-1)
							}

							if numTilesY > 1 {
								spacingY = (availableHeight - float64(numTilesY*height)) / float64(numTilesY-1)
							}

							for i := 0; i < numTilesY; i++ {
								posY := float64(i) * (float64(height) + spacingY)
								for j := 0; j < numTilesX; j++ {
									posX := float64(j) * (float64(width) + spacingX)
									can.DrawImage(resized, posX, posY)
								}
							}
						}
					} else if repeatX == "round" || repeatY == "round" {
						// Handle round - scales images to fit whole number of tiles
						if repeatX == "round" && repeatY != "round" {
							// Only horizontal rounding
							availableWidth := float64(canvasWidth)
							numTiles := math.Max(1, math.Round(availableWidth/float64(width)))
							roundedWidth := int(availableWidth / numTiles)

							// Create a new resized image with the rounded width
							roundedImg := image.NewRGBA(image.Rect(0, 0, roundedWidth, height))
							draw.CatmullRom.Scale(roundedImg, roundedImg.Bounds(), img, img.Bounds(), draw.Over, &draw.Options{})

							for i := 0; i < int(numTiles); i++ {
								posX := float64(i) * float64(roundedWidth)
								can.DrawImage(roundedImg, posX, float64(y))
							}
						} else if repeatX != "round" && repeatY == "round" {
							// Only vertical rounding
							availableHeight := float64(canvasHeight)
							numTiles := math.Max(1, math.Round(availableHeight/float64(height)))
							roundedHeight := int(availableHeight / numTiles)

							// Create a new resized image with the rounded height
							roundedImg := image.NewRGBA(image.Rect(0, 0, width, roundedHeight))
							draw.CatmullRom.Scale(roundedImg, roundedImg.Bounds(), img, img.Bounds(), draw.Over, &draw.Options{})

							for i := 0; i < int(numTiles); i++ {
								posY := float64(i) * float64(roundedHeight)
								can.DrawImage(roundedImg, float64(x), posY)
							}
						} else {
							// Both directions use round
							availableWidth := float64(canvasWidth)
							availableHeight := float64(canvasHeight)
							numTilesX := math.Max(1, math.Round(availableWidth/float64(width)))
							numTilesY := math.Max(1, math.Round(availableHeight/float64(height)))

							roundedWidth := int(availableWidth / numTilesX)
							roundedHeight := int(availableHeight / numTilesY)

							// Create a new resized image with the rounded dimensions
							roundedImg := image.NewRGBA(image.Rect(0, 0, roundedWidth, roundedHeight))
							draw.CatmullRom.Scale(roundedImg, roundedImg.Bounds(), img, img.Bounds(), draw.Over, &draw.Options{})

							for i := 0; i < int(numTilesY); i++ {
								posY := float64(i) * float64(roundedHeight)
								for j := 0; j < int(numTilesX); j++ {
									posX := float64(j) * float64(roundedWidth)
									can.DrawImage(roundedImg, posX, posY)
								}
							}
						}
					} else {
						// Default case: no-repeat in both directions
						can.DrawImage(resized, float64(x), float64(y))
					}
				} else {
					can.DrawImage(resized, float64(x), float64(y))
				}
			} else if len(bg.Image) > 18 && (bg.Image[0:16] == "linear-gradient(" || bg.Image[0:16] == "radial-gradient(") {
				// !ISSUE: GG fills rect completely with gradient
				var width, height int

				if bg.Size != "" {
					if bg.Size == "cover" {
						height = int(sHeight)
						width = int(sWidth)
					} else if bg.Size == "contain" || bg.Size == "auto" || bg.Size == "auto auto" {
						width = int(sWidth)
						height = int(sHeight)
					} else {
						parts := strings.Split(bg.Size, " ")
						if len(parts) == 2 {
							w := utils.ConvertToPixels(parts[0], self.EM, sWidth)
							h := utils.ConvertToPixels(parts[1], self.EM, sHeight)
							width = int(w)
							height = int(h)
						} else if len(parts) == 1 {
							s := utils.ConvertToPixels(parts[0], self.EM, sWidth)
							height = int(s) * (height / width)
							width = int(s)
						}
					}
				}
				x := 0
				y := 0

				if bg.Origin != "" {
					if bg.Origin == "padding-box" {
						x += int(self.Border.Left.Width)
						sW := int(sWidth - self.Border.Right.Width)
						sWidth = float32(sW)
						if width > sW {
							width = sW
						} else {
							width += x
						}

						y += int(self.Border.Top.Width)
						sH := int(sHeight - self.Border.Bottom.Width)
						sHeight = float32(sH)
						if height > sH {
							height = sH
						} else {
							height += y
						}
					} else if bg.Origin == "content-box" {
						x += int(self.Border.Left.Width + self.Padding.Left)
						sW := int(sWidth - (self.Border.Left.Width + self.Padding.Left + self.Border.Right.Width + self.Padding.Left))
						sWidth = float32(sW)
						if width > sW {
							width = sW
						}

						y += int(self.Border.Top.Width + self.Padding.Top)
						sH := int(sHeight - (self.Border.Top.Width + self.Padding.Top + self.Border.Bottom.Width + self.Padding.Top))
						sHeight = float32(sH)
						if height > sH {
							height = sH
						}
					}
				}

				if bg.PositionX != "" {
					parts := strings.Split(bg.PositionX, " ")

					if len(parts) == 2 {
						if parts[0] == "left" {
							d := int(utils.ConvertToPixels(parts[1], self.EM, self.Width))
							x += d
						} else if parts[0] == "right" {
							d := int(utils.ConvertToPixels(parts[1], self.EM, self.Width))
							x += int(sWidth) - (width + d)

						}
					} else if len(parts) == 1 {
						if parts[0] == "right" {
							x += int(sWidth) - (width)

						} else if parts[0] == "center" {
							x += int(sWidth/2) - (width / 2)
						} else {
							d := int(utils.ConvertToPixels(parts[0], self.EM, self.Width))
							x += d
						}
					}
				}

				if bg.PositionY != "" {
					parts := strings.Split(bg.PositionY, " ")

					if len(parts) == 2 {
						if parts[0] == "top" {
							d := int(utils.ConvertToPixels(parts[1], self.EM, self.Height))
							y += d
						} else if parts[0] == "bottom" {
							d := int(utils.ConvertToPixels(parts[1], self.EM, self.Height))
							y += int(sHeight) - (height + d)
						}
					} else if len(parts) == 1 {
						if parts[0] == "bottom" {
							y += int(sHeight) - (height)

						} else if parts[0] == "center" {
							y += int(sHeight/2) - (height / 2)
						} else {
							d := int(utils.ConvertToPixels(parts[0], self.EM, self.Height))
							y += d
						}
					}
				}
				var steps []step
				var lg gg.Gradient

				if bg.Image[0:16] == "linear-gradient(" {
					pg := parseLinearGradient(int(self.Width), int(self.Height), self.EM, bg.Image)
					// !NOTE: Does not support interpolation
					lg = can.CreateLinearGradient(pg.x1, pg.y1, pg.x2, pg.y2)
					steps = pg.steps
				} else if bg.Image[0:16] == "radial-gradient(" {
					pg := parseRadialGradient(int(sWidth), int(sHeight), self.EM, bg.Image)
					lg = can.CreateRadialGradient(pg.x1, pg.y1, pg.r1, pg.x2, pg.y2, pg.r2)
					steps = pg.steps
				}

				lg.AddColorStop(0, steps[0].color)
				for _, v := range steps {
					lg.AddColorStop(v.offset, v.color)
				}

				can.Context.SetFillStyle(lg)

				if bg.Repeat != "" {
					parts := strings.Split(bg.Repeat, " ")
					canvasHeight := self.Height + self.Border.Top.Width + self.Border.Bottom.Width
					canvasWidth := self.Width + self.Border.Left.Width + self.Border.Right.Width

					// Set default values for horizontal and vertical repeat
					repeatX := "no-repeat"
					repeatY := "no-repeat"

					// Handle single-value repeat
					if len(parts) == 1 {
						if parts[0] == "repeat-x" {
							repeatX = "repeat"
						} else if parts[0] == "repeat-y" {
							repeatY = "repeat"
						} else if parts[0] == "repeat" || parts[0] == "space" || parts[0] == "round" {
							repeatX = parts[0]
							repeatY = parts[0]
						}
					} else if len(parts) == 2 {
						repeatX = parts[0]
						repeatY = parts[1]
					}

					// Handle repeat-x and repeat-y as x/y directional repeats
					if repeatX == "repeat" && repeatY == "repeat" {
						// Calculate proper starting point to maintain pattern alignment
						tilesBackX := int(math.Ceil(float64(x) / float64(width)))
						tilesBackY := int(math.Ceil(float64(y) / float64(height)))

						startX := float64(x) - float64(tilesBackX*width)
						startY := float64(y) - float64(tilesBackY*height)

						// Draw covering the entire canvas
						for currentY := startY; currentY < float64(canvasHeight); currentY += float64(height) {
							for currentX := startX; currentX < float64(canvasWidth); currentX += float64(width) {
								can.FillRect(currentX, currentY, float64(width), float64(height))
							}
						}

					} else if repeatX == "repeat" && repeatY == "no-repeat" {
						// Repeat only horizontally at the specified y
						tilesBackX := int(math.Ceil(float64(x) / float64(width)))
						startX := float64(x) - float64(tilesBackX*width)
						startY := float64(y)

						for currentX := startX; currentX < float64(canvasWidth); currentX += float64(width) {
							can.FillRect(currentX, startY, float64(width), float64(height))
						}
					} else if repeatX == "no-repeat" && repeatY == "repeat" {
						// Repeat only vertically at the specified x
						tilesBackY := int(math.Ceil(float64(y) / float64(height)))
						startX := float64(x)
						startY := float64(y) - float64(tilesBackY*height)
						for currentY := startY; currentY < float64(canvasHeight); currentY += float64(height) {
							can.FillRect(startX, currentY, float64(width), float64(height))
						}
					} else if repeatX == "space" || repeatY == "space" {
						// Handle space - tiles with equal spacing, no clipping
						if repeatX == "space" && repeatY != "space" {
							// Only horizontal spacing
							availableWidth := float64(canvasWidth)
							numTiles := int(math.Floor(availableWidth / float64(width)))

							if numTiles > 1 {
								spacing := (availableWidth - float64(numTiles*width)) / float64(numTiles-1)
								for i := 0; i < numTiles; i++ {
									posX := float64(i) * (float64(width) + spacing)
									can.FillRect(posX, float64(y), float64(width), float64(height))
								}
							} else {
								// Only one tile - center it
								posX := (availableWidth - float64(width)) / 2
								can.FillRect(posX, float64(y), float64(width), float64(height))
							}
						} else if repeatX != "space" && repeatY == "space" {
							// Only vertical spacing
							availableHeight := float64(canvasHeight)
							numTiles := int(math.Floor(availableHeight / float64(height)))

							if numTiles > 1 {
								spacing := (availableHeight - float64(numTiles*height)) / float64(numTiles-1)
								for i := 0; i < numTiles; i++ {
									posY := float64(i) * (float64(height) + spacing)
									can.FillRect(float64(x), posY, float64(width), float64(height))
								}
							} else {
								// Only one tile - center it
								posY := (availableHeight - float64(height)) / 2
								can.FillRect(float64(x), posY, float64(width), float64(height))
							}
						} else {
							// Both directions use space
							availableWidth := float64(canvasWidth)
							availableHeight := float64(canvasHeight)
							numTilesX := int(math.Floor(availableWidth / float64(width)))
							numTilesY := int(math.Floor(availableHeight / float64(height)))

							spacingX := 0.0
							spacingY := 0.0

							if numTilesX > 1 {
								spacingX = (availableWidth - float64(numTilesX*width)) / float64(numTilesX-1)
							}

							if numTilesY > 1 {
								spacingY = (availableHeight - float64(numTilesY*height)) / float64(numTilesY-1)
							}

							for i := 0; i < numTilesY; i++ {
								posY := float64(i) * (float64(height) + spacingY)
								for j := 0; j < numTilesX; j++ {
									posX := float64(j) * (float64(width) + spacingX)
									can.FillRect(posX, posY, float64(width), float64(height))
								}
							}
						}
					} else if repeatX == "round" || repeatY == "round" {
						// Handle round - scales images to fit whole number of tiles
						if repeatX == "round" && repeatY != "round" {
							// Only horizontal rounding
							availableWidth := float64(canvasWidth)
							numTiles := math.Max(1, math.Round(availableWidth/float64(width)))
							roundedWidth := int(availableWidth / numTiles)

							for i := 0; i < int(numTiles); i++ {
								posX := float64(i) * float64(roundedWidth)
								can.FillRect(posX, float64(y), float64(roundedWidth), float64(height))
							}
						} else if repeatX != "round" && repeatY == "round" {
							// Only vertical rounding
							availableHeight := float64(canvasHeight)
							numTiles := math.Max(1, math.Round(availableHeight/float64(height)))
							roundedHeight := int(availableHeight / numTiles)

							for i := 0; i < int(numTiles); i++ {
								posY := float64(i) * float64(roundedHeight)
								can.FillRect(float64(x), posY, float64(width), float64(roundedHeight))
							}
						} else {
							// Both directions use round
							availableWidth := float64(canvasWidth)
							availableHeight := float64(canvasHeight)
							numTilesX := math.Max(1, math.Round(availableWidth/float64(width)))
							numTilesY := math.Max(1, math.Round(availableHeight/float64(height)))

							roundedWidth := int(availableWidth / numTilesX)
							roundedHeight := int(availableHeight / numTilesY)

							for i := 0; i < int(numTilesY); i++ {
								posY := float64(i) * float64(roundedHeight)
								for j := 0; j < int(numTilesX); j++ {
									posX := float64(j) * float64(roundedWidth)
									can.FillRect(posX, posY, float64(roundedWidth), float64(roundedHeight))
								}
							}
						}
					} else {
						// Default case: no-repeat in both directions
						can.FillRect(float64(x), float64(y), float64(width), float64(height))
					}
				} else {
					can.FillRect(float64(x), float64(y), float64(width), float64(height))
				}

			} else if len(bg.Image) > 18 && bg.Image[0:16] == "radial-gradient(" {
				parseRadialGradient(int(self.Width), int(self.Height), self.EM, bg.Image)
			}
		}

	}
	return can.Context.Image()
}

type LinearGradient struct {
	x1    float64
	y1    float64
	x2    float64
	y2    float64
	steps []step
}

type step struct {
	color  ic.RGBA
	offset float64
}

// background-image: linear-gradient(45deg, blue, red);
func parseLinearGradient(width, height int, em float32, lg string) LinearGradient {
	lg = strings.TrimPrefix(lg, "linear-gradient(")
	lg = strings.TrimSuffix(lg, ")")

	parts := element.Token('(', ')', ',', lg)
	var x1, y1, x2, y2 float64
	var col ic.RGBA
	var dist float32
	steps := []step{}
	for i, v := range parts {
		isAng := false
		if i == 0 {
			var angle string
			switch v {
			case "to top":
				angle = "0"
				isAng = true
			case "to bottom":
				angle = "180"
				isAng = true
			case "to left":
				angle = "270"
				isAng = true
			case "to right":
				angle = "90"
				isAng = true
			default:
				if strings.HasSuffix(v, "deg") {
					angle = strings.TrimSuffix(v, "deg")
					isAng = true
				} else {
					angle = "180"
				}
			}

			ang, _ := strconv.Atoi(angle)
			x1, y1, x2, y2 = calculateGradientPoints(float64(width), float64(height), float64(ang))
			dist = float32(math.Sqrt(((x2 - x1) * (x2 - x1)) + ((y2 - y1) * (y2 - y1))))
		}

		if !isAng {
			p := strings.Split(v, " ")

			c, err := color.ParseRGBA(p[0])
			if err == nil {
				col = c
				if len(p) == 1 {
					// just red
					steps = append(steps, step{
						color: col,
					})
					continue
				} else {
					// red 10% 20% etc..
					p = p[1:]

					for _, q := range p {
						steps = append(steps, step{
							color:  col,
							offset: float64(utils.ConvertToPixels(q, em, dist) / dist),
						})
					}
				}
			} else {
				// just 10%
				steps = append(steps, step{
					color:  col,
					offset: float64(utils.ConvertToPixels(v, em, dist) / dist),
				})
			}
		}
	}
	// also clean up ones that dont have a set distance
	// If you don't specify the location of a color, it is placed halfway between the one that precedes it and the one that follows it. The following two gradients are equivalent.
	var prevoffset, nextoffset float64
	for i, v := range steps {
		if v.offset == 0 {
			for c := i; c < len(steps); c++ {
				if steps[c].offset != 0 {
					nextoffset = steps[c].offset
					break
				}
			}
			if nextoffset == 0 {
				steps[i].offset = float64((dist/float32(len(steps)-1))*float32(i)) / float64(dist)
			} else {
				steps[i].offset = (nextoffset - prevoffset) / 2
			}
			prevoffset = steps[i].offset
		}
	}

	return LinearGradient{
		x1:    x1,
		y1:    y1,
		x2:    x2,
		y2:    y2,
		steps: steps,
	}
}

// CalculateGradientPoints takes a width, height, and CSS degree angle
// and returns the start and end points for a linear gradient
func calculateGradientPoints(width, height float64, cssAngle float64) (float64, float64, float64, float64) {
	// Convert CSS angle to radians (CSS angles are measured clockwise from the top)
	// We need to convert to mathematical angles (counterclockwise from right)
	angle := math.Pi * (360 - cssAngle) / 180

	// Find center of rectangle
	centerX := width / 2
	centerY := height / 2

	var startX, startY, endX, endY float64

	startX = centerX + (-height * math.Cos(angle))
	endX = centerX + (height * math.Cos(angle))
	startY = centerY + (-height * math.Sin(angle))
	endY = centerY + (height * math.Sin(angle))

	return startX, startY, endX, endY
}

type RadialGradient struct {
	x1    float64
	y1    float64
	r1    float64
	x2    float64
	y2    float64
	r2    float64
	steps []step
}

// background-image: linear-gradient(45deg, blue, red);
func parseRadialGradient(width, height int, em float32, rg string) RadialGradient {
	rg = strings.TrimPrefix(rg, "radial-gradient(")
	rg = strings.TrimSuffix(rg, ")")

	parts := element.Token('(', ')', ',', rg)

	var x1, y1, r1,
		x2, y2, r2 float64

	shape := "circle"
	size := "farthest-corner"
	position := "center"

	var col ic.RGBA
	var dist float32

	steps := []step{}

	// Parse the first argument
	isColorStop := true
	vs := strings.Split(parts[0], " ")
	atAt := false
	for _, p := range vs {
		if p == "at" {
			atAt = true
			continue
		}
		if atAt {
			//  <position>
			position = p
		} else {
			if p == "circle" || p == "ellipse" {
				// <radial-shape>
				shape = p
				isColorStop = false
			} else if p == "closest-corner" || p == "closest-side" ||
				p == "farthest-corner" || p == "farthest-side" {
				// <radial-extent>
				size = p
			} else {
				// <length> || <length-percentage>
				_, err := color.ParseRGBA(vs[0]) // make sure p is not a color
				if err != nil {
					size = p
				}
			}
		}
	}

	// !ISSUE: ellipitcal gradients are not supported
	// + instead match on both for circle (need to update gg)
	if shape == "circle" || shape == "ellipse" {
		ps := strings.Split(position, " ")
		if len(ps) == 2 {
			x1 = getPosition(ps[0], width, height, em, "x")
			y1 = getPosition(ps[1], width, height, em, "y")
		} else if len(ps) == 1 {
			x1 = getPosition(ps[0], width, height, em, "x")
			y1 = getPosition(ps[0], width, height, em, "y")
		} else {
			x1 = float64(width) / 2
			y1 = float64(height) / 2
		}

		// Set xy2 to be the same as xy1 to prevent a shift
		x2 = x1
		y2 = y1

		// Only r2 gets a size (it detirmines the size of the gradient)
		r2 = getSize(size, float64(width), float64(height), em, x1, y1)
		dist = float32(r2)
	}

	for i, v := range parts {
		vs = strings.Split(v, " ")

		if i == 0 && !isColorStop {
			continue
		}
		c, err := color.ParseRGBA(vs[0])
		if err == nil {
			col = c
			if len(vs) == 1 {
				// just red
				steps = append(steps, step{
					color: col,
				})
				continue
			} else {
				// red 10% 20% etc..
				vs = vs[1:]

				for _, q := range vs {
					steps = append(steps, step{
						color:  col,
						offset: float64(utils.ConvertToPixels(q, em, dist) / dist),
					})
				}
			}
		} else {
			// just 10%
			steps = append(steps, step{
				color:  col,
				offset: float64(utils.ConvertToPixels(v, em, dist) / dist),
			})
		}
	}

	// Interpolate the color steps
	var prevoffset, nextoffset float64
	for i, v := range steps {
		if v.offset == 0 {
			for c := i; c < len(steps); c++ {
				if steps[c].offset != 0 {
					nextoffset = steps[c].offset
					break
				}
			}
			if nextoffset == 0 {
				steps[i].offset = float64((dist/float32(len(steps)-1))*float32(i)) / float64(dist)
			} else {
				steps[i].offset = (nextoffset - prevoffset) / 2
			}
			prevoffset = steps[i].offset
		}
	}

	return RadialGradient{
		x1:    x1,
		y1:    y1,
		r1:    r1,
		x2:    x2,
		y2:    y2,
		r2:    r2,
		steps: steps,
	}
}

func getPosition(position string, width, height int, em float32, part string) float64 {
	switch position {
	case "left":
		return 0
	case "right":
		return float64(width)
	case "top":
		return 0
	case "bottom":
		return float64(height)
	case "center":
		if part == "x" {
			return float64(width) / 2
		} else if part == "y" {
			return float64(height) / 2
		}
	default:
		if part == "x" {
			return float64(utils.ConvertToPixels(position, em, float32(width)))
		} else if part == "y" {
			return float64(utils.ConvertToPixels(position, em, float32(height)))
		}
	}
	return 0
}

func getSize(size string, width, height float64, em float32, x, y float64) float64 {
	// Top left
	c1dist := math.Sqrt(((x - 0) * (x - 0)) + ((y - 0) * (y - 0)))
	// Top right
	c2dist := math.Sqrt(((x - width) * (x - width)) + ((y - 0) * (y - 0)))
	// Bottom right
	c3dist := math.Sqrt(((x - width) * (x - width)) + ((y - height) * (y - height)))
	// Bottom left
	c4dist := math.Sqrt(((x - 0) * (x - 0)) + ((y - height) * (y - height)))
	// Top
	s1dist := y
	// Right
	s2dist := math.Abs(x - width)
	// Bottom
	s3dist := math.Abs(y - height)
	// Left
	s4dist := x
	var res float64
	switch size {
	case "farthest-corner":
		if c1dist >= c2dist && c1dist >= c3dist && c1dist >= c4dist {
			res = c1dist
		} else if c2dist >= c1dist && c2dist >= c3dist && c2dist >= c4dist {
			res = c2dist
		} else if c3dist >= c1dist && c3dist >= c2dist && c3dist >= c4dist {
			res = c3dist
		} else if c4dist >= c1dist && c4dist >= c2dist && c4dist >= c3dist {
			res = c4dist
		}
	case "closest-corner":
		if c1dist <= c2dist && c1dist <= c3dist && c1dist <= c4dist {
			res = c1dist
		} else if c2dist <= c1dist && c2dist <= c3dist && c2dist <= c4dist {
			res = c2dist
		} else if c3dist <= c1dist && c3dist <= c2dist && c3dist <= c4dist {
			res = c3dist
		} else if c4dist <= c1dist && c4dist <= c2dist && c4dist <= c3dist {
			res = c4dist
		}
	case "farthest-side":
		if s1dist >= s2dist && s1dist >= s3dist && s1dist >= s4dist {
			res = s1dist
		} else if s2dist >= s1dist && s2dist >= s3dist && s2dist >= s4dist {
			res = s2dist
		} else if s3dist >= s1dist && s3dist >= s2dist && s3dist >= s4dist {
			res = s3dist
		} else if s4dist >= s1dist && s4dist >= s2dist && s4dist >= s3dist {
			res = s4dist
		}
	case "closest-side":
		if s1dist <= s2dist && s1dist <= s3dist && s1dist <= s4dist {
			res = s1dist
		} else if s2dist <= s1dist && s2dist <= s3dist && s2dist <= s4dist {
			res = s2dist
		} else if s3dist <= s1dist && s3dist <= s2dist && s3dist <= s4dist {
			res = s3dist
		} else if s4dist <= s1dist && s4dist <= s2dist && s4dist <= s3dist {
			res = s4dist
		}
	default:
		// !NOTE: This is the corrent way to do it for circles not ellipses (but ellipses aren't supported)
		if width < height {
			res = float64(utils.ConvertToPixels(size, em, float32(width)))
		} else {
			res = float64(utils.ConvertToPixels(size, em, float32(height)))
		}
	}
	return res
}
