package cstyle

import (
	"fmt"
	adapter "gui/adapters"
	"gui/border"
	"gui/color"
	"gui/element"
	"gui/font"
	"gui/library"
	"gui/parser"
	"gui/selector"
	"gui/utils"
	"image"
	"os"
	"sort"
	"strconv"
	"strings"

	imgFont "golang.org/x/image/font"
)

type Plugin struct {
	Selector func(*element.Node) bool
	Level    int
	Handler  func(*element.Node, *map[string]element.State)
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
	Options      adapter.Options
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
	dat, _ := os.ReadFile(path)
	styles, styleMaps := parser.ParseCSS(string(dat))

	if c.StyleMap == nil {
		c.StyleMap = map[string][]*parser.StyleMap{}
	}

	for k, v := range styleMaps {
		if c.StyleMap[k] == nil {
			c.StyleMap[k] = []*parser.StyleMap{}
		}
		for styleMapKey := range v {
			v[styleMapKey].SheetNumber = len(c.StyleSheets)
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
		for styleMapKey := range v {
			v[styleMapKey].SheetNumber = len(c.StyleSheets)
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

	// !IDEA: Might be able to only reload page if element that is being hoverved over has a possible :hover class
	// + might addeventlisteners here?????

	// Apply styles from style sheets
	selectors := []string{}

	selectors = append(selectors, n.TagName)

	if n.Id != "" {
		selectors = append(selectors, "#"+n.Id)
	}

	for _, class := range n.ClassList.Classes {
		if class[0] == ':' {
			selectors = append(selectors, class)
		} else {
			selectors = append(selectors, "."+class)
		}
	}

	styleMaps := []*parser.StyleMap{}
	for _, v := range selectors {
		styleMaps = append(styleMaps, c.StyleMap[v]...)
	}

	for _, styleMap := range styleMaps {
		parts := styleMap.Selector
		currentElement := n
		isPseudo := false
		pseudoSelector := ""
		for i, part := range parts {
			for i, v := range part {
				if len(v) < 2 {
					continue
				}
				if v[0:2] == "::" {
					part = append(part[:i], part[i+1:]...)
					isPseudo = true
					pseudoSelector = v
					break
				}
			}

			has := selector.Contains(part, selectors)

			if i == len(parts)-1 || !has {
				if has {
					if isPseudo {
						for k, v := range *styleMap.Styles {
							if pseudoStyles[pseudoSelector] == nil {
								pseudoStyles[pseudoSelector] = map[string]string{}
							}
							pseudoStyles[pseudoSelector][k] = v
						}

					} else {
						for k, v := range *styleMap.Styles {
							styles[k] = v
						}
					}
				}
				break
			} else {
				selectors = []string{}
				currentElement = currentElement.Parent

				for _, class := range currentElement.ClassList.Classes {
					if class[0] == ':' {
						selectors = append(selectors, class)
					} else {
						selectors = append(selectors, "."+class)
					}
				}

				if n.Id != "" {
					selectors = append(selectors, "#"+currentElement.Id)
				}

				selectors = append(selectors, currentElement.TagName)

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
		if parentZIndex, ok := n.Parent.Style["z-index"]; ok && parentZIndex != "" {
			z, _ := strconv.Atoi(parentZIndex)
			z += 1
			styles["z-index"] = strconv.Itoa(z)
		}
	}

	return styles, pseudoStyles
}

func (c *CSS) AddPlugin(plugin Plugin) {
	plugins := c.Plugins
	plugins = append(plugins, plugin)
	sort.Slice(plugins, func(i, j int) bool {
		return plugins[i].Level < plugins[j].Level
	})
	c.Plugins = plugins
}

func (c *CSS) AddTransformer(transformer Transformer) {
	c.Transformers = append(c.Transformers, transformer)
}

func (c *CSS) ComputeNodeStyle(n *element.Node, state *map[string]element.State, shelf *library.Shelf) *element.Node {
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
			y = (base.Height - height) - utils.ConvertToPixels(bottomVal, self.EM, parent.Width)
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
		fmt.Println(n.TagName, n.InnerText)
		n.InnerText = strings.TrimSpace(n.InnerText)
		self = genTextNode(n, state, c, shelf)
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
	var childYOffset float32

	for i := 0; i < len(n.Children); i++ {
		v := n.Children[i]
		v.Parent = n
		n.Children[i] = c.ComputeNodeStyle(v, state, shelf)
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
			self.ScrollHeight = sh
		}

		if cState.Width > self.Width {
			self.Width = cState.Width
		}
	}

	if style["height"] == "" {
		self.Height += self.Padding.Bottom
	}

	self.ScrollHeight += int(self.Padding.Bottom)

	(*state)[n.Properties.Id] = self

	for _, v := range plugins {
		if v.Selector(n) {
			v.Handler(n, state)
		}
	}

	if n.Properties.Id == "input7" {
		fmt.Println(n.Properties.Id, self.Width, self.Border)
	}

	return n
}

func genTextNode(n *element.Node, state *map[string]element.State, css *CSS, shelf *library.Shelf) element.State {
	s := *state
	self := s[n.Properties.Id]
	parent := s[n.Parent.Properties.Id]

	self.Textures = []string{}

	text := element.Text{}

	italic := false

	if n.Style["font-weight"] == "bold" {
		n.Style["font-weight"] = "700"
	}
	if n.Style["font-weight"] == "bolder" {
		n.Style["font-weight"] = "900"
	}
	if n.Style["font-weight"] == "lighter" {
		n.Style["font-weight"] = "200"
	}
	if n.Style["font-weight"] == "normal" {
		n.Style["font-weight"] = "400"
	}

	if n.Style["font-style"] == "italic" {
		italic = true
	}

	if css.Fonts == nil {
		css.Fonts = map[string]imgFont.Face{}
	}
	fid := n.Style["font-family"] + fmt.Sprint(self.EM, n.Style["font-weight"], italic)
	if css.Fonts[fid] == nil {
		weight, _ := strconv.Atoi(n.Style["font-weight"])
		f, _ := font.LoadFont(n.Style["font-family"], int(self.EM), weight, italic)
		css.Fonts[fid] = f
	}
	fnt := css.Fonts[fid]
	text.Font = &fnt

	letterSpacing := utils.ConvertToPixels(n.Style["letter-spacing"], self.EM, parent.Width)
	wordSpacing := utils.ConvertToPixels(n.Style["word-spacing"], self.EM, parent.Width)
	lineHeight := utils.ConvertToPixels(n.Style["line-height"], self.EM, parent.Width)

	if lineHeight == 0 {
		lineHeight = self.EM + 3
	}

	text.LineHeight = int(lineHeight)
	text.WordSpacing = int(wordSpacing)
	text.LetterSpacing = int(letterSpacing)
	wb := " "

	if n.Style["word-wrap"] == "break-word" {
		wb = ""
	}

	if n.Style["text-wrap"] == "wrap" || n.Style["text-wrap"] == "balance" {
		wb = ""
	}

	var dt float32

	if n.Style["text-decoration-thickness"] == "auto" || n.Style["text-decoration-thickness"] == "" {
		dt = self.EM / 7
	} else {
		dt = utils.ConvertToPixels(n.Style["text-decoration-thickness"], self.EM, parent.Width)
	}

	col := color.Parse(n.Style, "font")

	if n.Style["text-decoration-color"] == "" {
		n.Style["text-decoration-color"] = n.Style["color"]
	}

	// self.Color = col

	text.Color = col
	text.DecorationColor = color.Parse(n.Style, "decoration")
	text.Align = n.Style["text-align"]
	text.WordBreak = wb
	text.WordSpacing = int(wordSpacing)
	text.LetterSpacing = int(letterSpacing)
	text.WhiteSpace = n.Style["white-space"]
	text.DecorationThickness = int(dt)
	text.Overlined = n.Style["text-decoration"] == "overline"
	text.Underlined = n.Style["text-decoration"] == "underline"
	text.LineThrough = n.Style["text-decoration"] == "linethrough"
	text.EM = int(self.EM)
	text.Width = int(parent.Width)
	text.Text = n.InnerText
	// text.Last = n.GetAttribute("last") == "true"

	if n.Style["word-spacing"] == "" {
		text.WordSpacing = font.MeasureSpace(&text)
	}
	key := text.Text + utils.RGBAtoString(text.Color) + utils.RGBAtoString(text.DecorationColor) + text.Align + text.WordBreak + strconv.Itoa(text.WordSpacing) + strconv.Itoa(text.LetterSpacing) + text.WhiteSpace + strconv.Itoa(text.DecorationThickness) + strconv.Itoa(text.EM)
	key += strconv.FormatBool(text.Overlined) + strconv.FormatBool(text.Underlined) + strconv.FormatBool(text.LineThrough)

	exists := shelf.Check(key)
	var width int
	if exists {
		lookup := make(map[string]struct{}, len(self.Textures))
		for _, v := range self.Textures {
			lookup[v] = struct{}{}
		}

		if _, found := lookup[key]; !found {
			self.Textures = append(self.Textures, key)
		}
		width = font.MeasureText(&text, text.Text+" ")
	} else {
		var data *image.RGBA
		data, width = font.Render(&text)
		self.Textures = append(self.Textures, shelf.Set(key, data))
	}

	if n.Style["height"] == "" && n.Style["min-height"] == "" {
		self.Height = float32(text.LineHeight)
	}

	if n.Style["width"] == "" && n.Style["min-width"] == "" {
		self.Width = float32(width)
	}

	return self
}
