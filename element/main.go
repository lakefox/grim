package element

import (
	"bytes"
	"fmt"
	"grim/canvas"
	ic "image/color"
	"slices"
	"strings"
)

type Node struct {
	TagName           string
	InnerText         string
	Parent            *Node
	Children          []*Node
	style             map[string]string
	ComputedStyle     map[string]string
	Id                string
	ClassList         ClassList
	Href              string
	Src               string
	Title             string
	attribute         map[string]string
	ScrollLeft        int
	ScrollTop         int
	TabIndex          int
	ContentEditable   bool
	Required          bool
	Disabled          bool
	Checked           bool
	Focused           bool
	Hovered           bool
	InitalStyles      map[string]string
	StyleSheets       *Styles
	ConditionalStyles map[string]map[string]string

	// !NOTE: ScrollHeight is the amount of scroll left, not the total amount of scroll
	// + if you  want the same scrollHeight like js the add the height of the element to it
	ScrollHeight   int
	ScrollWidth    int
	Canvas         *canvas.Canvas
	PseudoElements map[string]map[string]string

	Value         string
	OnClick       func(Event)
	OnContextMenu func(Event)
	OnMouseDown   func(Event)
	OnMouseUp     func(Event)
	OnMouseEnter  func(Event)
	OnMouseLeave  func(Event)
	OnMouseOver   func(Event)
	OnMouseMove   func(Event)
	OnScroll      func(Event)
	Properties    Properties
}

// !TODO: I would like to remove element.Node.Properties if possible but I don't think it is
type Properties struct {
	Id             string
	EventListeners map[string][]func(Event)
	// Events         []string
	// !TODO: Make selected work
	Selected []float32
}

type State struct {
	X               float32
	Y               float32
	Z               float32
	Width           float32
	Height          float32
	Border          Border
	Textures        map[string]string
	EM              float32
	Background      ic.RGBA
	Margin          BoxSpacing
	Padding         BoxSpacing
	Cursor          string
	Crop            Crop
	Hidden          bool
	ScrollHeight    int
	ScrollWidth     int
	ContentEditable bool
	Value           string
	TabIndex        int
}

type Crop struct {
	X      int
	Y      int
	Width  int
	Height int
}

type BoxSpacing struct {
	Top    float32
	Left   float32
	Right  float32
	Bottom float32
}

type Background struct {
	Color      ic.RGBA
	Image      string
	PositionX  string
	PositionY  string
	Size       string
	Repeat     string
	Origin     string
	Clip       string
	Attachment string
}

// !MAN: Node Style getter/setter
// + [!DEVMAN]Note: Contains all user inputed styles, all inline styles over ride stylesheet styles
func (n *Node) SetStyle(key, value string) {
	n.style[key] = value
}

func (n *Node) GetStyle(key string) string {
	return n.style[key]
}

func (n *Node) Styles() map[string]string {
	return n.style
}

func (n *Node) GetComputedStyle(key string) string {
	return n.ComputedStyle[key]
}

// !MAN: Generates the InnerHTML of an element
// !TODO: Add a setter
func (n *Node) InnerHTML() string {
	// !TODO: Will need to update the styles once this can be parsed
	return InnerHTML(n)
}

// !MAN: Generates the OuterHTML of an element
// !TODO: Add a setter
func (n *Node) OuterHTML() string {
	tag, closing := NodeToHTML(n)
	return tag + InnerHTML(n) + closing
}

type ClassList struct {
	classes []string
}

// !ISSUE: Need to getstyles here
// + try to use conditional styles
func (c *ClassList) Add(class string) {
	if !slices.Contains(c.classes, class) {
		c.classes = append(c.classes, class)
	}
}

func (c *ClassList) Remove(class string) {
	for i, v := range c.classes {
		if v == class {
			c.classes = append(c.classes[:i], c.classes[i+1:]...)
			break
		}
	}
}

func (c *ClassList) Classes(value ...[]string) []string {
	if len(value) > 0 {
		for _, v := range value {
			c.classes = append(c.classes, v...)
		}
	}
	return c.classes
}

type Border struct {
	Top    BorderSide
	Right  BorderSide
	Bottom BorderSide
	Left   BorderSide
	Radius BorderRadius
}

type BorderSide struct {
	Width float32
	Style string
	Color ic.RGBA
}

type BorderRadius struct {
	TopLeft     float32
	TopRight    float32
	BottomLeft  float32
	BottomRight float32
}

// !MAN: Attribute getter
// + name: string containing the attribute name
// + [!DEVMAN]Note: If you want to get all of the attributes use the .attribute prop (only for element)
func (n *Node) GetAttribute(name string) string {
	return n.attribute[name]
}

func (n *Node) SetAttribute(key, value string) {
	n.attribute[key] = value
	if n.Parent != nil {
		n.StyleSheets.GetStyles(n)
	}
}

func (n *Node) CreateElement(name string) Node {
	ti := -1

	focusableElements := map[string]bool{
		"input":    true,
		"button":   true,
		"select":   true,
		"textarea": true,
		"output":   true,
		"a":        true,
		"area":     true,
		"audio":    true,
		"video":    true,
		"details":  true,
		"label":    true,
	}

	if focusableElements[name] {
		ti = 9999999
	}
	return Node{
		TagName:   name,
		InnerText: "",
		Children:  []*Node{},
		Id:        "",
		ClassList: ClassList{
			classes: []string{},
		},
		Href:              "",
		Src:               "",
		Title:             "",
		attribute:         make(map[string]string),
		ComputedStyle:     make(map[string]string),
		ConditionalStyles: make(map[string]map[string]string),
		style:             make(map[string]string),
		Value:             "",
		TabIndex:          ti,
		ContentEditable:   false,
		StyleSheets:       n.StyleSheets,
		Properties: Properties{
			Id:             "",
			EventListeners: make(map[string][]func(Event)),
			Selected:       []float32{},
		},
	}
}

func Itoa(num int) string {
	if num == 0 {
		return "0"
	}

	isNegative := num < 0
	if isNegative {
		num = -num
	}

	dic := [10]byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}
	result := make([]byte, 0, 10) // Preallocate a slice for better performance

	for num > 0 {
		rv := num % 10
		result = append(result, dic[rv])
		num /= 10
	}

	// Reverse the result slice since digits are appended in reverse order
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	if isNegative {
		return "-" + string(result)
	}
	return string(result)
}

func GenerateUniqueId(parent *Node, tagName string) string {
	return parent.Properties.Id + ":" + tagName + Itoa(len(parent.Children))
}

func (n *Node) AppendChild(c *Node) {
	c.Parent = n
	c.Properties.Id = GenerateUniqueId(n, c.TagName)
	n.Children = append(n.Children, c)
	if n.Parent != nil {
		n.StyleSheets.GetStyles(c)
	}
}

func (n *Node) InsertAfter(c, tgt *Node) {
	c.Parent = n
	c.Properties.Id = GenerateUniqueId(n, c.TagName)
	if n.Parent != nil {
		n.StyleSheets.GetStyles(c)
	}
	nodeIndex := -1
	for i, v := range n.Children {
		if v.Properties.Id == tgt.Properties.Id {
			nodeIndex = i
			break
		}
	}
	if nodeIndex > -1 {
		n.Children = append(n.Children[:nodeIndex+1], append([]*Node{c}, n.Children[nodeIndex+1:]...)...)
	} else {
		n.AppendChild(c)
	}
}

func (n *Node) InsertBefore(c, tgt *Node) {
	c.Parent = n
	// Set Id

	c.Properties.Id = GenerateUniqueId(n, c.TagName)
	if n.Parent != nil {
		n.StyleSheets.GetStyles(c)
	}
	nodeIndex := -1
	for i, v := range n.Children {
		if v.Properties.Id == tgt.Properties.Id {
			nodeIndex = i
			break
		}
	}

	if nodeIndex > 0 {
		n.Children = append(n.Children[:nodeIndex], append([]*Node{c}, n.Children[nodeIndex:]...)...)
	} else {
		n.Children = append([]*Node{c}, n.Children...)
	}

}

func (n *Node) Remove() {
	nodeIndex := -1
	for i, v := range n.Parent.Children {
		if v.Properties.Id == n.Properties.Id {
			nodeIndex = i
			break
		}
	}
	if nodeIndex > 0 {
		n.Parent.Children = append(n.Parent.Children[:nodeIndex], n.Parent.Children[nodeIndex+1:]...)
	}
	n.Parent.StyleSheets.GetStyles(n.Parent)
}

func (n *Node) Focus() {
	n.Focused = true
	f := n.ConditionalStyles[":focus"]
	if f != nil {
		for k, v := range f {
			n.ComputedStyle[k] = v
		}
	}
}

func (n *Node) Blur() {
	n.Focused = false
	h := n.ConditionalStyles[":hover"]
	f := n.ConditionalStyles[":focus"]
	if f != nil {
		for k := range f {
			if n.Hovered {
				if h[k] != "" {
					n.ComputedStyle[k] = h[k]
				} else {
					n.ComputedStyle[k] = n.InitalStyles[k]
				}
			} else {
				n.ComputedStyle[k] = n.InitalStyles[k]
			}
		}
	}
}

func (n *Node) GetContext(width, height int) *canvas.Canvas {
	n.ComputedStyle["width"] = Itoa(width) + "px"
	n.ComputedStyle["height"] = Itoa(height) + "px"
	ctx := canvas.NewCanvas(width, height)
	n.Canvas = ctx
	return ctx
}

func (n *Node) ScrollTo(x, y int) {
	n.ScrollTop = y
	n.ScrollLeft = x
}

type Event struct {
	X           int
	Y           int
	KeyCode     int
	ScrollX     int
	ScrollY     int
	Key         string
	CtrlKey     bool
	MetaKey     bool
	ShiftKey    bool
	AltKey      bool
	Click       bool
	ContextMenu bool
	MouseDown   bool
	MouseUp     bool
	MouseEnter  bool
	MouseLeave  bool
	MouseOver   bool
	MouseMove   bool
	KeyUp       bool
	KeyDown     bool
	KeyPress    bool
	Input       bool
	Target      *Node
	Name        string
	Data        any
	Value       string
	Hover       bool
}

type EventList struct {
	Event Event
	List  []string
}

func (node *Node) AddEventListener(name string, callback func(Event)) {
	if node.Properties.EventListeners == nil {
		node.Properties.EventListeners = make(map[string][]func(Event))
	}
	if node.Properties.EventListeners[name] == nil {
		node.Properties.EventListeners[name] = []func(Event){}
	}
	if !funcInSlice(callback, node.Properties.EventListeners[name]) {
		node.Properties.EventListeners[name] = append(node.Properties.EventListeners[name], callback)
	}
}

func (node *Node) DispatchEvent(event Event) {
	for _, v := range node.Properties.EventListeners[event.Name] {
		v(event)
	}
}

func funcInSlice(f func(Event), slice []func(Event)) bool {
	for _, item := range slice {
		// Compare function values directly
		if fmt.Sprintf("%p", item) == fmt.Sprintf("%p", f) {
			return true
		}
	}
	return false
}

func NodeToHTML(node *Node) (string, string) {
	// if node.TagName == "text" {
	// 	return node.InnerText + " ", ""
	// }

	var buffer bytes.Buffer
	buffer.WriteString("<" + node.TagName)

	if node.ContentEditable {
		buffer.WriteString(" contentEditable=\"true\"")
	}

	// Add ID if present
	if node.Id != "" {
		buffer.WriteString(" id=\"" + node.Id + "\"")
	}

	// Add ID if present
	if node.Title != "" {
		buffer.WriteString(" title=\"" + node.Title + "\"")
	}

	// Add ID if present
	if node.Src != "" {
		buffer.WriteString(" src=\"" + node.Src + "\"")
	}

	// Add ID if present
	if node.Href != "" {
		buffer.WriteString(" href=\"" + node.Href + "\"")
	}

	// Add class list if present
	if len(node.ClassList.Classes()) > 0 {
		classes := ""
		for _, v := range node.ClassList.Classes() {
			if len(v) > 0 {
				if string(v[0]) != ":" {
					classes += v + " "
				}
			}
		}
		classes = strings.TrimSpace(classes)
		if len(classes) > 0 {
			buffer.WriteString(" class=\"" + classes + "\"")
		}
	}

	styles := node.Styles()

	// Add style if present
	if len(styles) > 0 {

		style := ""
		for key, value := range styles {
			if key != "inlineText" {
				style += key + ":" + value + ";"
			}
		}
		style = strings.TrimSpace(style)

		if len(style) > 0 {
			buffer.WriteString(" style=\"" + style + "\"")
		}
	}

	attributeString := ""
	for k, v := range node.attribute {
		attributeString += " " + k + "=\"" + v + "\""
	}
	// Add other attributes if present
	buffer.WriteString(attributeString)

	buffer.WriteString(">")

	// Add inner text if present
	if node.InnerText != "" && !ChildrenHaveText(node) {
		buffer.WriteString(node.InnerText)
	}
	return buffer.String(), "</" + node.TagName + ">"
}

func OuterHTML(node *Node) string {
	var buffer bytes.Buffer

	tag, closing := NodeToHTML(node)

	buffer.WriteString(tag)

	// Recursively add children
	for _, child := range node.Children {
		buffer.WriteString(OuterHTML(child))
	}

	buffer.WriteString(closing)

	return buffer.String()
}

func InnerHTML(node *Node) string {
	var buffer bytes.Buffer
	// Recursively add children
	for _, child := range node.Children {
		buffer.WriteString(OuterHTML(child))
	}
	return buffer.String()
}

func ChildrenHaveText(n *Node) bool {
	for _, child := range n.Children {
		if len(strings.TrimSpace(child.InnerText)) != 0 {
			return true
		}
		// Recursively check if any child nodes have text
		if ChildrenHaveText(child) {
			return true
		}
	}
	return false
}
