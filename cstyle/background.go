package cstyle

import (
	"bytes"
	"fmt"
	"golang.org/x/image/draw"
	"grim/canvas"
	"grim/color"
	"grim/element"
	"grim/utils"
	"image"
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
		s := strings.Split(style[v], ",")
		f := []string{}
		for _, v := range s {
			if strings.TrimSpace(v) != "" {
				f = append(f, v)
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
	fmt.Println("=======")
	for _, bg := range self.Background {
		fmt.Println(bg)
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
								fmt.Println(width, height)
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
				fmt.Println(width, height, sWidth, sHeight)
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
			} else if len(bg.Image) > 18 && bg.Image[0:16] == "linear-gradient(" {
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

				parseLinearGradient(int(self.Width), int(self.Height), bg.Image)
				// !NOTE: Does not support interpolation
				// lg := can.NewLinearGradient()
			}
		}

	}
	return can.Context.Image()
}

type LinearGradient struct {
	angle int
	steps []step
}

type step struct {
	color  image.RGBA
	length int
}

// background-image: linear-gradient(45deg, blue, red);
func parseLinearGradient(width, height int, lg string) LinearGradient {
	lg = strings.TrimPrefix(lg, "linear-gradient(")
	lg = strings.TrimSuffix(lg, ")")

	parts := element.Token('(', ')', ',', lg)

	for _, v := range parts {
		fmt.Println("PART", v)
	}

	return LinearGradient{}
}
