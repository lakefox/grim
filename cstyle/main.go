package cstyle

import (
	"fmt"
	adapter "grim/adapters"
	"grim/border"
	"grim/color"
	"grim/element"
	"grim/font"
	"grim/parser"
	"grim/utils"
	"image"
	"strconv"
	"strings"

	"golang.org/x/image/draw"

	imgFont "golang.org/x/image/font"
)

type Plugin struct {
	Selector func(*element.Node) bool
	Handler  func(*element.Node, *map[string]element.State, *CSS)
}

type Transformer struct {
	Selector func(*element.Node) bool
	Handler  func(*element.Node, *CSS) *element.Node
}

type CSS struct {
	Width        float32
	Height       float32
	StyleSheets  []map[string]*map[string]string
	Plugins      []Plugin
	Transformers []Transformer
	Document     *element.Node
	Fonts        map[string]imgFont.Face
	StyleMap     map[string][]*parser.StyleMap
	Adapter      *adapter.Adapter
	Path         string
}

func (c *CSS) Transform(n *element.Node) *element.Node {
	for _, v := range c.Transformers {
		if v.Selector(n) {
			n = v.Handler(n, c)
		}
	}

	for i := 0; i < len(n.Children); i++ {
		tc := c.Transform(n.Children[i])
		n.Children[i] = tc
	}

	return n
}

func (c *CSS) StyleSheet(path string) {
	// Parse the CSS file
	data, _ := c.Adapter.FileSystem.ReadFile(path)
	styles, styleMaps := parser.ParseCSS(string(data))

	if c.StyleMap == nil {
		c.StyleMap = map[string][]*parser.StyleMap{}
	}

	for k, v := range styleMaps {
		if c.StyleMap[k] == nil {
			c.StyleMap[k] = []*parser.StyleMap{}
		}
		c.StyleMap[k] = append(c.StyleMap[k], v...)
	}

	c.StyleSheets = append(c.StyleSheets, styles)
}

func (c *CSS) StyleTag(css string) {
	styles, styleMaps := parser.ParseCSS(css)

	if c.StyleMap == nil {
		c.StyleMap = map[string][]*parser.StyleMap{}
	}

	for k, v := range styleMaps {
		if c.StyleMap[k] == nil {
			c.StyleMap[k] = []*parser.StyleMap{}
		}
		c.StyleMap[k] = append(c.StyleMap[k], v...)
	}

	c.StyleSheets = append(c.StyleSheets, styles)
}

var inheritedProps = []string{
	"color",
	"cursor",
	"font",
	"font-family",
	"font-size",
	"font-style",
	"font-weight",
	"letter-spacing",
	"line-height",
	// "text-align",
	"text-indent",
	"text-justify",
	"text-shadow",
	"text-transform",
	"text-decoration",
	"visibility",
	"word-spacing",
	"display",
	"scrollbar-color",
}

func (c *CSS) QuickStyles(n *element.Node) map[string]string {
	styles := make(map[string]string)

	// Inherit styles from parent
	if n.Parent != nil {
		ps := n.Parent.Style
		for _, prop := range inheritedProps {
			if value, ok := ps[prop]; ok && value != "" {
				styles[prop] = value
			}
		}
	}

	// Add node's own styles
	for k, v := range n.Style {
		styles[k] = v
	}

	return styles
}

func (c *CSS) GetStyles(n *element.Node) (map[string]string, map[string]map[string]string) {
	styles := make(map[string]string)
	pseudoStyles := make(map[string]map[string]string)

	// Inherit styles from parent
	if n.Parent != nil {
		ps := n.Parent.Style
		for _, prop := range inheritedProps {
			if value, ok := ps[prop]; ok && value != "" {
				styles[prop] = value
			}
		}
	}

	// Add node's own styles
	for k, v := range n.Style {
		styles[k] = v
	}

	baseSelectors := element.GenBaseElements(n)
	parentSelectors := element.GenBaseElements(n.Parent)

	for _, v := range baseSelectors {
		sm := c.StyleMap[v]
		for _, m := range sm {
			if element.ShouldTestSelector(n, m.Selector) {
				match, isPseudo := element.TestSelector(n, m.Selector)
				if match {
					if isPseudo {
						pseudoSelector := "::" + strings.Split(m.Selector, "::")[1]
						if pseudoStyles[pseudoSelector] == nil {
							pseudoStyles[pseudoSelector] = map[string]string{}
						}
						for k, v := range *m.Styles {
							if pseudoStyles[pseudoSelector] == nil {
								pseudoStyles[pseudoSelector] = map[string]string{}
							}
							pseudoStyles[pseudoSelector][k] = v
						}
					} else {
						for k, v := range *m.Styles {
							styles[k] = v
						}
					}
				}
			}
		}
	}

	for _, v := range parentSelectors {
		sm := c.StyleMap[v]
		for _, m := range sm {
			if strings.Contains(m.Selector, ":nth-child(") {
				if element.ShouldTestSelector(n, m.Selector) {
					match, _ := element.TestSelector(n, m.Selector)
					if match {
						for k, v := range *m.Styles {
							styles[k] = v
						}
					}
				}
			}
		}
	}

	// fmt.Println(n.Properties.Id, styles, n.Parent.Properties.Id, n.Parent.Style)

	// 	if has {
	// 		if isPseudo {
	// 			for k, v := range *styleMap.Styles {
	// 				if pseudoStyles[pseudoSelector] == nil {
	// 					pseudoStyles[pseudoSelector] = map[string]string{}
	// 				}
	// 				pseudoStyles[pseudoSelector][k] = v
	// 			}
	// 		} else {
	// 			for k, v := range *styleMap.Styles {
	// 				styles[k] = v
	// 			}
	// 		}
	// 	}
	// }

	// Parse inline styles
	inlineStyles := parser.ParseStyleAttribute(n.GetAttribute("style"))
	for k, v := range inlineStyles {
		styles[k] = v
	}

	// Handle z-index inheritance
	if n.Parent != nil && styles["z-index"] == "" {
		if parentZIndex, ok := n.Parent.Style["z-index"]; ok && parentZIndex != "" {
			z, _ := strconv.Atoi(parentZIndex)
			z += 1
			styles["z-index"] = strconv.Itoa(z)
		}
	}

	return styles, pseudoStyles
}

func (c *CSS) AddPlugin(plugin Plugin) {
	c.Plugins = append(c.Plugins, plugin)
}

func (c *CSS) AddTransformer(transformer Transformer) {
	c.Transformers = append(c.Transformers, transformer)
}

func (c *CSS) ComputeNodeStyle(n *element.Node, state *map[string]element.State) *element.Node {
	shelf := c.Adapter.Library
	// Head is not renderable
	if utils.IsParent(*n, "head") {
		return n
	}

	s := *state
	self := s[n.Properties.Id]
	plugins := c.Plugins
	parent := s[n.Parent.Properties.Id]

	// Cache the style map
	style := n.Style
	self.Background = color.Parse(style, "background")
	self.Border, _ = border.Parse(style, self, parent)
	border.Draw(&self, shelf)

	fs := utils.ConvertToPixels(style["font-size"], parent.EM, parent.Width)
	self.EM = fs

	if style["display"] == "none" {
		self.X, self.Y, self.Width, self.Height = 0, 0, 0, 0
		(*state)[n.Properties.Id] = self
		return n
	}

	// Set Z index value to be sorted in window
	if zIndex, err := strconv.Atoi(style["z-index"]); err == nil {
		self.Z = float32(zIndex)
	}
	if parent.Z > 0 {
		self.Z = parent.Z + 1
	}

	(*state)[n.Properties.Id] = self

	wh := utils.GetWH(*n, state)
	width, height := wh.Width, wh.Height

	x, y := parent.X, parent.Y
	offsetX, offsetY := utils.GetXY(n, state)
	x += offsetX
	y += offsetY

	m := utils.GetMP(*n, wh, state, "margin")
	p := utils.GetMP(*n, wh, state, "padding")
	self.Margin = m
	self.Padding = p
	self.Cursor = n.Style["cursor"]

	var top, left, right, bottom bool

	if style["position"] == "absolute" {
		bas := utils.GetPositionOffsetNode(n.Parent)
		base := s[bas.Properties.Id]
		if topVal := style["top"]; topVal != "" {
			y = utils.ConvertToPixels(topVal, self.EM, parent.Width) + base.Y
			top = true
		}
		if leftVal := style["left"]; leftVal != "" {
			x = utils.ConvertToPixels(leftVal, self.EM, parent.Width) + base.X
			left = true
		}
		if rightVal := style["right"]; rightVal != "" {
			x = base.X + ((base.Width - width) - utils.ConvertToPixels(rightVal, self.EM, parent.Width))
			right = true
		}
		if bottomVal := style["bottom"]; bottomVal != "" {
			y = base.Y + ((base.Height - height) - utils.ConvertToPixels(bottomVal, self.EM, parent.Width))
			bottom = true
		}
	} else {
		for i, v := range n.Parent.Children {
			if v.Style["position"] != "absolute" {
				if v.Properties.Id == n.Properties.Id {
					if i > 0 {
						sib := n.Parent.Children[i-1]
						sibling := s[sib.Properties.Id]
						if sib.Style["position"] != "absolute" {
							if style["display"] == "inline" {
								y = sibling.Y
								if sib.Style["display"] != "inline" {
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

	self.Width = width
	self.Height = height

	self.ContentEditable = n.ContentEditable

	(*state)[n.Properties.Id] = self

	if !utils.ChildrenHaveText(n) && len(n.InnerText) > 0 {
		n.InnerText = strings.TrimSpace(n.InnerText)
		italic := false

		if n.Style["font-style"] == "italic" {
			italic = true
		}

		if c.Fonts == nil {
			c.Fonts = map[string]imgFont.Face{}
		}
		fid := n.Style["font-family"] + fmt.Sprint(self.EM, n.Style["font-weight"], italic)

		if c.Fonts[fid] == nil {
			f, _ := font.LoadFont(n.Style["font-family"], int(self.EM), n.Style["font-weight"], italic, &c.Adapter.FileSystem)
			c.Fonts[fid] = f
		}
		fnt := c.Fonts[fid]

		metadata := font.GetMetaData(n, state, &fnt)
		key := font.Key(metadata)
		exists := c.Adapter.Library.Check(key)
		var width int
		if exists {
			lookup := make(map[string]struct{}, len(self.Textures))
			for _, v := range self.Textures {
				lookup[v] = struct{}{}
			}

			if _, found := lookup[key]; !found {
				self.Textures = append(self.Textures, key)
			}
			width, _ = font.MeasureText(metadata, metadata.Text+" ")
		} else {
			var data *image.RGBA
			data, width = font.Render(metadata)
			self.Textures = append(self.Textures, c.Adapter.Library.Set(key, data))
		}

		if n.Style["height"] == "" && n.Style["min-height"] == "" {
			self.Height = float32(metadata.LineHeight)
		}

		if n.Style["width"] == "" && n.Style["min-width"] == "" {
			self.Width = float32(width)
		}
	}

	// Load canvas into textures
	if n.TagName == "canvas" {
		if n.Canvas != nil {
			found := false
			key := n.Properties.Id + "canvas"
			for _, v := range self.Textures {
				if v == key {
					found = true
				}
			}

			if n.Canvas.RGBA.Bounds().Dx() != int(self.Width) || n.Canvas.RGBA.Bounds().Dy() != int(self.Height) {
				resized := image.NewRGBA(image.Rect(0, 0, int(self.Width), int(self.Height)))
				draw.CatmullRom.Scale(resized, resized.Bounds(), n.Canvas.RGBA, n.Canvas.RGBA.Bounds(), draw.Over, &draw.Options{})
				n.Canvas.RGBA = resized
			}

			can := shelf.Set(key, n.Canvas.RGBA)
			if !found {
				self.Textures = append(self.Textures, can)
			}
		}
	}

	self.Value = n.InnerText
	self.TabIndex = n.TabIndex
	(*state)[n.Properties.Id] = self
	(*state)[n.Parent.Properties.Id] = parent

	self.ScrollHeight = 0
	self.ScrollWidth = 0
	var childYOffset float32

	for i := 0; i < len(n.Children); i++ {
		v := n.Children[i]
		v.Parent = n
		n.Children[i] = c.ComputeNodeStyle(v, state)
		cState := (*state)[n.Children[i].Properties.Id]

		if style["height"] == "" && style["max-height"] == "" {
			if v.Style["position"] != "absolute" && cState.Y+cState.Height > childYOffset {
				childYOffset = cState.Y + cState.Height
				self.Height = cState.Y - self.Border.Top.Width - self.Y + cState.Height
				self.Height += cState.Margin.Top + cState.Margin.Bottom + cState.Padding.Top + cState.Padding.Bottom + cState.Border.Top.Width + cState.Border.Bottom.Width
			}
		}
		sh := int((cState.Y + cState.Height) - self.Y)
		if self.ScrollHeight < sh {
			if n.Children[i].TagName != "grim-track" {
				self.ScrollHeight = sh
			}
		}

		sw := int((cState.X + cState.Width) - self.X)
		if self.ScrollWidth < sw {
			if n.Children[i].TagName != "grim-track" {
				self.ScrollWidth = sw
			}
		}

		if cState.Width > self.Width && n.Style["width"] == "" {
			self.Width = cState.Width
		}
	}

	if style["height"] == "" {
		self.Height += self.Padding.Bottom
	}
	self.ScrollHeight += int(self.Padding.Bottom)
	self.ScrollWidth += int(self.Padding.Right)

	(*state)[n.Properties.Id] = self

	for _, v := range plugins {
		if v.Selector(n) {
			v.Handler(n, state, c)
		}
	}
	// if n.TagName != "notaspan" {
	// 	fmt.Println(">>>", n.Properties.Id, self.X, self.Y, self.Width, self.Height)
	// 	for k, v := range style {
	// 		fmt.Println(k, v)
	// 	}
	// }

	return n
}
