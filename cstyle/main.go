package cstyle

import (
	"fmt"
	"github.com/golang/freetype/truetype"
	adapter "grim/adapters"
	"grim/border"
	"grim/element"
	"grim/font"
	"grim/utils"
	"image"
	"strconv"
	"strings"

	"golang.org/x/image/draw"
)

type Plugin struct {
	Selector func(*element.Node, *CSS) bool
	Handler  func(*element.Node, *CSS)
}

type Transformer struct {
	Selector func(*element.Node, *CSS) bool
	Handler  func(*element.Node, *CSS) *element.Node
}

type CSS struct {
	Width        float32
	Height       float32
	Plugins      []Plugin
	Transformers []Transformer
	Fonts        map[string]*truetype.Font
	Adapter      *adapter.Adapter
	Path         string
	State        map[string]element.State
}

func (c *CSS) AddPlugin(plugin Plugin) {
	c.Plugins = append(c.Plugins, plugin)
}

func (c *CSS) AddTransformer(transformer Transformer) {
	c.Transformers = append(c.Transformers, transformer)
}

var nonRenderTags = map[string]bool{
	"head":  true,
	"meta":  true,
	"link":  true,
	"title": true,
	"style": true,
}

func (c *CSS) ComputeNodeStyle(n *element.Node) element.State {
	// Head is not renderable
	s := c.State
	self := s[n.Properties.Id]
	if self.Textures == nil {
		self.Textures = map[string]string{}
	}

	if nonRenderTags[n.TagName()] {
		return self
	}

	for _, v := range c.Transformers {
		if v.Selector(n, c) {
			v.Handler(n, c)
		}
	}

	plugins := c.Plugins
	parentNode := n.Parent()
	parent := s[parentNode.Properties.Id]
	// Cache the style map
	style := n.ComputedStyle

	for k, v := range n.Styles() {
		style[k] = v
	}

	self.Border, _ = border.Parse(style, self, parent)
	// Remove border if its 0
	if self.Border.Top.Width+self.Border.Right.Width+self.Border.Left.Width+self.Border.Bottom.Width == 0 {
		self.Textures["border"] = ""
	}

	if style["font-size"] == "" {
		n.ComputedStyle["font-size"] = "1em"
	}
	fs := utils.ConvertToPixels(n.ComputedStyle["font-size"], parent.EM, parent.Width)
	self.EM = fs

	if style["display"] == "none" {
		self.X, self.Y, self.Width, self.Height = 0, 0, 0, 0
		c.State[n.Properties.Id] = self
		return self
	}

	// Set Z index value to be sorted in window
	if zIndex, err := strconv.Atoi(style["z-index"]); err == nil {
		self.Z = float32(zIndex)
	}

	if self.Z > 0 {
		self.Z = parent.Z + 1
	}

	c.State[n.Properties.Id] = self

	self.Background = ParseBackground(style)

	c.State[n.Properties.Id] = self
	wh, m, p := utils.FindBounds(*n, style, &c.State)

	self.Margin = m
	self.Padding = p
	self.Width = wh.Width
	self.Height = wh.Height
	self.Cursor = style["cursor"]
	c.State[n.Properties.Id] = self

	x, y := parent.X, parent.Y
	offsetX, offsetY := utils.GetXY(*n, c.State)
	x += offsetX
	y += offsetY

	var top, left, right, bottom bool
	if style["position"] == "absolute" {
		// !DEVMAN: Properties.Id is the ancestory of an element with colons seperating them
		// + if we split them up we can check the parents without recursion or a while (for true) loop
		// + NOTE: See utils.GenerateUnqineId to see how they are made
		ancestors := strings.Split(n.Properties.Id, ":")

		offsetNode := n
		// Should skip the current element and the ROOT
		for i := len(ancestors) - 2; i > 0; i-- {
			offsetNode = offsetNode.Parent()
			pos := offsetNode.ComputedStyle["position"]
			if pos == "relative" || pos == "absolute" {
				break
			}
		}

		base := s[offsetNode.Properties.Id]
		if topVal := style["top"]; topVal != "" {
			y = utils.ConvertToPixels(topVal, self.EM, parent.Width) + base.Y
			top = true
		}
		if leftVal := style["left"]; leftVal != "" {
			x = utils.ConvertToPixels(leftVal, self.EM, parent.Width) + base.X
			left = true
		}
		if rightVal := style["right"]; rightVal != "" {
			x = base.X + ((base.Width - self.Width) - utils.ConvertToPixels(rightVal, self.EM, parent.Width))
			right = true
		}
		if bottomVal := style["bottom"]; bottomVal != "" {
			y = base.Y + ((base.Height - self.Height) - utils.ConvertToPixels(bottomVal, self.EM, parent.Width))
			bottom = true
		}
	} else {
		for i, v := range parentNode.Children {
			if v.ComputedStyle["position"] != "absolute" {
				if v.Properties.Id == n.Properties.Id {
					if i > 0 {
						sib := parentNode.Children[i-1]
						sibling := s[sib.Properties.Id]
						if sib.ComputedStyle["position"] != "absolute" {
							if style["display"] == "inline" {
								y = sibling.Y
								if sib.ComputedStyle["display"] != "inline" {
									y += sibling.Height
								}
							} else {
								y = sibling.Y + sibling.Height + sibling.Border.Top.Width + sibling.Border.Bottom.Width + sibling.Margin.Bottom
							}
						}
					}
					break
				} else if style["display"] != "inline" {
					vState := s[v.Properties.Id]
					y += vState.Margin.Top + vState.Margin.Bottom + vState.Padding.Top + vState.Padding.Bottom + vState.Height + self.Border.Top.Width
				}
			}
		}
	}

	relPos := !top && !left && !right && !bottom
	if left || relPos {
		x += m.Left
	}
	if top || relPos {
		y += m.Top
	}
	if right {
		x -= m.Right
	}
	if bottom {
		y -= m.Bottom
	}

	self.X = x
	self.Y = y

	self.ContentEditable = n.ContentEditable

	c.State[n.Properties.Id] = self
	innerText := n.GetInnerText()
	if !element.ChildrenHaveText(n) && len(innerText) > 0 {
		n.SetInnerText(strings.TrimSpace(innerText))
		italic := false

		if style["font-style"] == "italic" {
			italic = true
		}

		if c.Fonts == nil {
			c.Fonts = map[string]*truetype.Font{}
		}
		fid := style["font-family"] + fmt.Sprint(style["font-weight"], italic)
		fnt, ok := c.Fonts[fid]

		if !ok {
			f, err := font.LoadFont(style["font-family"], int(self.EM), style["font-weight"], italic, &c.Adapter.FileSystem)
			if err != nil {
				panic(err)
			}
			c.Fonts[fid] = f
			fnt = f
		}

		metadata := font.GetMetaData(n, style, &c.State, fnt)
		key := font.Key(metadata)
		m, exists := c.Adapter.Textures[n.Properties.Id]["text"]
		var width int

		if exists && m == key {
			width = font.MeasureText(metadata, metadata.Text+" ")
		} else {
			if exists {
				c.Adapter.UnloadTexture(n.Properties.Id, "text")
			}
			var data image.Image
			data, width = font.RenderFont(metadata)
			c.Adapter.LoadTexture(n.Properties.Id, "text", key, data)
		}
		self.Textures["text"] = key

		if (style["height"] == "" && style["min-height"] == "") || n.TagName() == "text" {
			self.Height = float32(metadata.LineHeight)
			n.ComputedStyle["height"] = strconv.Itoa(int(self.Height)) + "px"
		}

		if style["width"] == "" && style["min-width"] == "" {
			self.Width = float32(width)
			n.ComputedStyle["width"] = strconv.Itoa(int(self.Width)) + "px"
		}
	}

	// Load canvas into textures
	if n.TagName() == "canvas" {
		if n.Canvas != nil {
			key := n.Properties.Id + "canvas"

			img := n.Canvas.Context.Image()
			b := img.Bounds()
			if b.Dx() != int(self.Width) || b.Dy() != int(self.Height) {
				resized := image.NewRGBA(image.Rect(0, 0, int(self.Width), int(self.Height)))
				draw.CatmullRom.Scale(resized, resized.Bounds(), img, img.Bounds(), draw.Over, &draw.Options{})
				// n.Canvas.RGBA = resized
			}

			c.Adapter.UnloadTexture(n.Properties.Id, "canvas")
			c.Adapter.LoadTexture(n.Properties.Id, "canvas", key, img)
			self.Textures["canvas"] = key
		}
	}

	self.Value = n.GetInnerText()
	self.TabIndex = n.TabIndex
	c.State[n.Properties.Id] = self
	c.State[parentNode.Properties.Id] = parent

	self.ScrollHeight = 0
	self.ScrollWidth = 0
	var childYOffset float32

	for i := 0; i < len(n.Children); i++ {
		cState := c.ComputeNodeStyle(n.Children[i])

		if style["height"] == "" && style["max-height"] == "" {
			if n.Children[i].ComputedStyle["position"] != "absolute" && cState.Y+cState.Height > childYOffset {
				childYOffset = cState.Y + cState.Height
				self.Height = cState.Y - self.Border.Top.Width - self.Y + cState.Height
				self.Height += cState.Margin.Top + cState.Margin.Bottom + cState.Padding.Top + cState.Padding.Bottom + cState.Border.Top.Width + cState.Border.Bottom.Width
			}
		}

		sh := int((cState.Y + cState.Height) - self.Y)
		if self.ScrollHeight < sh {
			if n.Children[i].TagName() != "grim-track" {
				self.ScrollHeight = sh
				self.ScrollHeight += int(cState.Margin.Top + cState.Margin.Bottom + cState.Padding.Top + cState.Padding.Bottom + cState.Border.Top.Width + cState.Border.Bottom.Width)

			}
		}

		sw := int((cState.X + cState.Width) - self.X)

		if self.ScrollWidth < sw {
			if n.Children[i].TagName() != "grim-track" {
				self.ScrollWidth = sw
			}
		}
		if cState.ScrollWidth > self.ScrollWidth {
			self.ScrollWidth = cState.ScrollWidth
		}

		if cState.Width > self.Width && style["width"] == "" {
			self.Width = cState.Width
		}
	}

	self.ScrollHeight += int(self.Padding.Bottom + self.Padding.Top)
	self.ScrollWidth += int(self.Padding.Right)
	if style["height"] == "" {
		self.Height += self.Padding.Bottom
	}
	c.State[n.Properties.Id] = self

	border.Draw(&self, c.Adapter, n.Properties.Id)
	c.State[n.Properties.Id] = self

	for _, v := range plugins {
		if v.Selector(n, c) {
			v.Handler(n, c)
		}
	}
	self = c.State[n.Properties.Id]
	return self
}
