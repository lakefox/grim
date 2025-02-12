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
	"sort"
	"strconv"
	"strings"

	"golang.org/x/image/draw"

	imgFont "golang.org/x/image/font"
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
	Document     map[string]*element.Node
	Fonts        map[string]imgFont.Face
	StyleMap     map[string][]*parser.StyleMap
	Adapter      *adapter.Adapter
	Path         string
	StyleSheets  int
	State        map[string]element.State
	PsuedoStyles map[string]map[string]map[string]string
}

func (c *CSS) StyleSheet(path string) {
	// Parse the CSS file
	data, _ := c.Adapter.FileSystem.ReadFile(path)
	styleMaps := parser.ParseCSS(string(data), c.StyleSheets)
	c.StyleSheets++

	if c.StyleMap == nil {
		c.StyleMap = map[string][]*parser.StyleMap{}
	}

	for k, v := range styleMaps {
		if c.StyleMap[k] == nil {
			c.StyleMap[k] = []*parser.StyleMap{}
		}
		c.StyleMap[k] = append(c.StyleMap[k], v...)
	}
}

func (c *CSS) StyleTag(css string) {
	styleMaps := parser.ParseCSS(css, c.StyleSheets)
	c.StyleSheets++

	if c.StyleMap == nil {
		c.StyleMap = map[string][]*parser.StyleMap{}
	}

	for k, v := range styleMaps {
		if c.StyleMap[k] == nil {
			c.StyleMap[k] = []*parser.StyleMap{}
		}
		c.StyleMap[k] = append(c.StyleMap[k], v...)
	}
}

var inheritedProps = []string{
	"color",
	"cursor",
	"font",
	"font-family",
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
		ps := n.Parent.Styles()
		for _, prop := range inheritedProps {
			if value, ok := ps[prop]; ok && value != "" {
				styles[prop] = value
			}
		}
	}

	// Add node's own styles
	for k, v := range n.Styles() {
		styles[k] = v
	}

	return styles
}

// !ISSUE: GetStyles only needs to be ran if a new node is added, and the inital run, or a style tag innerHTML chanages
// + rest can be done with a modified QuickStyles
// + kinda see that note for a complete list

func (c *CSS) GetStyles(n *element.Node) {
	styles := make(map[string]string)
	pseudoStyles := make(map[string]map[string]string)

	// Inherit styles from parent
	if n.Parent != nil {
		ps := n.Parent.Styles() 
		for _, prop := range inheritedProps {
			if value, ok := ps[prop]; ok && value != "" {
				styles[prop] = value
			}
		}
	}

	// Add node's own styles
	for k, v := range n.Styles() {
		styles[k] = v
	}

	baseSelectors := element.GenBaseElements(n)
	testedSelectors := map[string]bool{}

	// !DEVMAN: You need to pre-sort the selectors by their .Sheet field to create the
	// + cascading effect of CSS

	styleMaps := []*parser.StyleMap{}
	for _, v := range baseSelectors {
		sm := c.StyleMap[v]
		styleMaps = append(styleMaps, sm...)
	}
	sort.Slice(styleMaps, func(i, j int) bool {
		return styleMaps[i].Sheet < styleMaps[j].Sheet
	})
	for _, m := range styleMaps {
		if element.ShouldTestSelector(n, m.Selector) {
			testedSelectors[m.Selector] = true
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

	// Parse inline styles
	inlineStyles := parser.ParseStyleAttribute(n.GetAttribute("style"))
	for k, v := range inlineStyles {
		styles[k] = v
	}

	// Handle z-index inheritance
	if n.Parent != nil && styles["z-index"] == "" {
		parentZIndex := n.Parent.Style("z-index")
		if parentZIndex != "" {
			z, _ := strconv.Atoi(parentZIndex)
			z += 1
			styles["z-index"] = strconv.Itoa(z)
		}
	}

	for k, v := range styles {
		n.Style(k,v)
	}	
	c.PsuedoStyles[n.Properties.Id] = pseudoStyles
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
	shelf := c.Adapter.Library
	// Head is not renderable
	s := c.State
	self := s[n.Properties.Id]

	if nonRenderTags[n.TagName] {
		return self
	}

	for _, v := range c.Transformers {
		if v.Selector(n, c) {
			v.Handler(n, c)
		}

		// if st, ok := c.Styles[n.Properties.Id]; ok {
		// 	for k,v  := range st {
		// 		n.Style(k,v)
		// 	}
		// }
	}

	plugins := c.Plugins
	parent := s[n.Parent.Properties.Id]

	// Cache the style map
	style := map[string]string{}	
	//
	// Map added styles to the style object
	for k, v := range n.Styles() {
		style[k] = v
	}

	self.Background, _ = color.Parse(style, "background")
	self.Border, _ = border.Parse(style, self, parent)

	if style["font-size"] == "" {
		n.Style("font-size", "1em")
	}
	fs := utils.ConvertToPixels(n.Style("font-size"), parent.EM, parent.Width)
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

	c.State[n.Properties.Id] = self
	wh, m, p := utils.FindBounds(*n, style, c.State)

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
		// + if we split them up we can check the parents without recusion or a while (for true) loop
		// + NOTE: See utils.GenerateUnqineId to see how they are made
		ancestors := strings.Split(n.Properties.Id, ":")

		offsetNode := n
		// Should skip the current element and the ROOT
		for i := len(ancestors) - 2; i > 0; i-- {
			offsetNode = offsetNode.Parent
			pos := offsetNode.Style("position")
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
		for i, v := range n.Parent.Children {
			if v.Style("position") != "absolute" {
				if v.Properties.Id == n.Properties.Id {
					if i > 0 {
						sib := n.Parent.Children[i-1]
						sibling := s[sib.Properties.Id]
						if sib.Style("position") != "absolute" {
							if style["display"] == "inline" {
								y = sibling.Y
								if sib.Style("display") != "inline" {
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

	if !utils.ChildrenHaveText(n) && len(n.InnerText) > 0 {
		n.InnerText = strings.TrimSpace(n.InnerText)
		italic := false

		if style["font-style"] == "italic" {
			italic = true
		}

		if c.Fonts == nil {
			c.Fonts = map[string]imgFont.Face{}
		}
		fid := style["font-family"] + fmt.Sprint(self.EM, style["font-weight"], italic)

		if c.Fonts[fid] == nil {
			f, err := font.LoadFont(style["font-family"], int(self.EM), style["font-weight"], italic, &c.Adapter.FileSystem)

			if err != nil {
				panic(err)
			}
			c.Fonts[fid] = f
		}

		fnt := c.Fonts[fid]

		metadata := font.GetMetaData(n, style, &c.State, &fnt)
		key := font.Key(metadata)
		exists := c.Adapter.Library.Check(key)
		// var width int
		if exists {
			lookup := make(map[string]struct{}, len(self.Textures))
			for _, v := range self.Textures {
				lookup[v] = struct{}{}
			}

			if _, found := lookup[key]; !found {
				self.Textures = append(self.Textures, key)
			}
			// width, _ = font.MeasureText(metadata, metadata.Text+" ")
		} else {
			var data *image.RGBA
			data, _ = font.Render(metadata)
			self.Textures = append(self.Textures, c.Adapter.Library.Set(key, data))
		}

		if style["height"] == "" && style["min-height"] == "" {
			self.Height = float32(metadata.LineHeight)
		}

		// if style["width"] == "" && style["min-width"] == "" {
		// 	self.Width = float32(width)
		// }
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
	c.State[n.Properties.Id] = self
	c.State[n.Parent.Properties.Id] = parent

	self.ScrollHeight = 0
	self.ScrollWidth = 0
	var childYOffset float32
	
	for i := 0; i < len(n.Children); i++ {
		v := n.Children[i]
		v.Parent = n
		cState := c.ComputeNodeStyle(v)

		if style["height"] == "" && style["max-height"] == "" {
			if v.Style("position") != "absolute" && cState.Y+cState.Height > childYOffset {
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

		if cState.Width > self.Width && style["width"] == "" {
			self.Width = cState.Width
		}
	}

	if style["height"] == "" {
		self.Height += self.Padding.Bottom
	}
	self.ScrollHeight += int(self.Padding.Bottom)
	self.ScrollWidth += int(self.Padding.Right)

	c.State[n.Properties.Id] = self

	border.Draw(&self, shelf)
	c.State[n.Properties.Id] = self

	for _, v := range plugins {
		if v.Selector(n, c) {
			v.Handler(n, c)
		}
	}
	
	return self
}
