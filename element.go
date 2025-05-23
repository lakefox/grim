package grim

import (
	"bytes"
	"fmt"
	"grim/canvas"
	ic "image/color"
	"slices"
	"strconv"
	"strings"
)

// !TODO: Make everything a setter
// + if things become slow (they should) due to the appendchild, add a compute pause function that skips all check for the internal running,
// + can add to ComputeNodeState at the start and then removed at the end
type Node struct {
	tagName           string            // non modifiable
	innerText         string            // modifable
	parent            *Node             // nm
	Children          []*Node           // m
	style             styles // nm
	id                string            // m
	ClassList         ClassList
	href              string // m
	src               string // m
	title             string // m
	attribute         map[string]string
	scrollLeft        int                          // m
	scrollTop         int                          // m
	tabIndex          int                          // m
	contentEditable   bool                         // m
	required          bool                         // m
	disabled          bool                         // m
	checked           bool                         // m
	focused           bool                         // nm
	hovered           bool                         // nm
	StyleSheets       *Styles                      // nm

	// !NOTE: ScrollHeight is the amount of scroll left, not the total amount of scroll
	// + if you  want the same scrollHeight like js the add the height of the element to it
	scrollHeight   int // nm
	scrollWidth    int // nm
	Canvas         *canvas.Canvas

	value         string // m
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

type styles {
	// In order, all CSS style sections that apply to this node
	styles []map[string]string
	// Styles the user set to the element and styles set internally
	inline map[string]string
	// Styles that sometimes apply and can be computed on init render
	// like :focus and :hover
	conditional map[string]map[string]string
	// Styles for pseudo elements associated with the node
	pseudo map[string]map[string]string
}


// This function is used to build the styles from scratch
func (n *Node) GetComputedStyles() map[string]string {
	styles := map[string]string{}

	// !OPT: This will be slow
	// Inherit styles from the parent
	if n.parent != nil {
		ps := n.parent.GetComputedStyles()
		for _, prop := range inheritedProps {
			if value, ok := ps[prop]; ok && value != "" {
				styles[prop] = value
			}
		}
	}

	// Casscade the styles from the stylesheets
	for _, s := range n.style.styles {
		for k, v := range s {
			styles[k] = v
		}
	}

	// Then apply the conditional styles
	if n.hovered && n.styles.conditional[":hover"] != nil {
		for k, v := range n.styles.conditional[":hover"] {
			styles[k] = v
		}
	}
	if n.focused && n.styles.conditional[":focus"] != nil {
		for k, v := range n.styles.conditional[":focus"] {
			styles[k] = v
		}
	}

	for k, v := range n.styles.inline {
		styles[k] = v
	}

	return styles
}

// !TODO: I would like to remove element.Node.Properties if possible but I don't think it is possible
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
	Background      []Background
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
	Attachment string
}

func (n *Node) TagName() string {
	return n.tagName
}

func (n *Node) Parent() *Node {
	return n.parent
}

func (n *Node) InnerText(value ...string) string {
	if len(value) != 0 {
		n.innerText = value[0]
	}
	return n.innerText
}

func (n *Node) Id(value ...string) string {
	if len(value) != 0 {
		n.id = value[0]
	}
	return n.id
}

func (n *Node) Href(value ...string) string {
	if len(value) != 0 {
		n.href = value[0]
	}
	return n.href
}

func (n *Node) Src(value ...string) string {
	if len(value) != 0 {
		n.src = value[0]
	}
	return n.src
}

func (n *Node) Title(value ...string) string {
	if len(value) != 0 {
		n.title = value[0]
	}
	return n.title
}

func (n *Node) ContentEditable(value ...bool) bool {
	if len(value) != 0 {
		n.contentEditable = value[0]
	}
	return n.contentEditable
}

func (n *Node) TabIndex(value ...int) int {
	if len(value) != 0 {
		n.tabIndex = value[0]
	}
	return n.tabIndex
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
func (n *Node) GetOuterHTML() string {
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
	if n.parent != nil {
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
		tagName:   name,
		innerText: "",
		Children:  []*Node{},
		id:        "",
		ClassList: ClassList{
			classes: []string{},
		},
		href:              "",
		src:               "",
		title:             "",
		attribute:         make(map[string]string),
		ComputedStyle:     make(map[string]string),
		ConditionalStyles: make(map[string]map[string]string),
		style:             make(map[string]string),
		value:             "",
		tabIndex:          ti,
		contentEditable:   false,
		StyleSheets:       n.StyleSheets,
		Properties: Properties{
			Id:             "",
			EventListeners: make(map[string][]func(Event)),
			Selected:       []float32{},
		},
	}
}

func GenerateUniqueId(parent *Node, tagName string) string {
	return parent.Properties.Id + ":" + tagName + strconv.Itoa(len(parent.Children))
}

func (n *Node) AppendChild(c *Node) {
	c.parent = n
	c.Properties.Id = GenerateUniqueId(n, c.tagName)
	n.Children = append(n.Children, c)
	if n.parent != nil {
		n.StyleSheets.GetStyles(c)
	}
}

func (n *Node) InsertAfter(c, tgt *Node) {
	c.parent = n
	c.Properties.Id = GenerateUniqueId(n, c.tagName)
	if n.parent != nil {
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
	c.parent = n
	// Set Id

	c.Properties.Id = GenerateUniqueId(n, c.tagName)
	if n.parent != nil {
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
	for i, v := range n.parent.Children {
		if v.Properties.Id == n.Properties.Id {
			nodeIndex = i
			break
		}
	}
	if nodeIndex > 0 {
		n.parent.Children = append(n.parent.Children[:nodeIndex], n.parent.Children[nodeIndex+1:]...)
	}
	n.parent.StyleSheets.GetStyles(n.parent)
}

func (n *Node) Focus() {
	n.focused = true
	ConditionalStyleHandler(n, map[string]string{})
}

func (n *Node) Blur() {
	n.focused = false
	ConditionalStyleHandler(n, map[string]string{})
}

func (n *Node) GetContext(width, height int) *canvas.Canvas {
	n.ComputedStyle["width"] = strconv.Itoa(width) + "px"
	n.ComputedStyle["height"] = strconv.Itoa(height) + "px"
	ctx := canvas.NewCanvas(width, height)
	n.Canvas = ctx
	return ctx
}

func (n *Node) ScrollTo(x, y int) {
	n.scrollLeft = x
	n.scrollTop = y
}

func (n *Node) ScrollBy(x, y int) {
	n.scrollTop += y
	n.scrollLeft += x
}

// Left, Right
func (n *Node) GetScroll() (int, int) {
	return n.scrollLeft, n.scrollTop
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
	var buffer bytes.Buffer
	buffer.WriteString("<" + node.tagName)

	if node.contentEditable {
		buffer.WriteString(" contentEditable=\"true\"")
	}

	// Add ID if present
	if node.id != "" {
		buffer.WriteString(" id=\"" + node.id + "\"")
	}

	// Add ID if present
	if node.title != "" {
		buffer.WriteString(" title=\"" + node.title + "\"")
	}

	// Add ID if present
	if node.src != "" {
		buffer.WriteString(" src=\"" + node.src + "\"")
	}

	// Add ID if present
	if node.href != "" {
		buffer.WriteString(" href=\"" + node.href + "\"")
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

	styles := node.style
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
	if node.innerText != "" && !ChildrenHaveText(node) {
		buffer.WriteString(node.innerText)
	}
	return buffer.String(), "</" + node.tagName + ">"
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
		if len(strings.TrimSpace(child.innerText)) != 0 {
			return true
		}
		// Recursively check if any child nodes have text
		if ChildrenHaveText(child) {
			return true
		}
	}
	return false
}

func CopyDocument(node *Node, parent *Node) *Node {
	n := *node
	n.parent = parent
	// !DEVMAN: Copying is done here, would like to remove this and add it to ComputeNodeStyle, so I can save a tree climb

	if len(node.Children) > 0 {
		n.Children = make([]*Node, 0, len(node.Children))
		for i := range node.Children {
			n.Children = append(n.Children, node.Children[i])
		}
		for i := range node.Children {
			n.Children[i] = CopyDocument(node.Children[i], &n)
		}

	}

	return &n
}
