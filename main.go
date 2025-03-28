package grim

import (
	_ "embed"
	"fmt"
	adapter "grim/adapters"
	"grim/cstyle"
	"grim/cstyle/plugins/crop"
	"grim/cstyle/plugins/flex"
	"grim/cstyle/plugins/inline"
	"grim/cstyle/plugins/textAlign"
	"grim/cstyle/transformers/banda"
	"time"

	marginblock "grim/cstyle/transformers/margin-block"
	"grim/cstyle/transformers/ol"
	"grim/cstyle/transformers/scrollbar"
	"grim/cstyle/transformers/text"
	"grim/cstyle/transformers/ul"
	"grim/element"
	"grim/events"
	"grim/font"
	"grim/scripts"
	"grim/scripts/a"
	"grim/utils"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/golang/freetype/truetype"
	"golang.org/x/net/html"
)

//go:embed master.css
var mastercss string

type Window struct {
	CSS        cstyle.CSS
	document   element.Node
	Styles     element.Styles
	Scripts    scripts.Scripts
	RenderData []element.State
	Rerender   bool
	shouldStop bool
}

func (w *Window) Document() *element.Node {
	return &w.document
}

// !TODO: Add a Mux option to all a http server to map to the window
func (window *Window) HttpMux() {}

func (window *Window) Path(path string) {
	styleSheets, styleTags, htmlNodes := parseHTMLFromFile(path, window.CSS.Adapter.FileSystem)

	for _, v := range styleSheets {
		data, _ := window.CSS.Adapter.FileSystem.ReadFile(v)
		window.Styles.StyleTag(string(data))
	}

	for _, v := range styleTags {
		window.Styles.StyleTag(v)
	}

	window.CSS.Path = filepath.Dir(path)
	createNode(htmlNodes, &window.document, &window.Styles)
	open(window)
}

func New(adapterFunction *adapter.Adapter, width, height int) Window {
	w := Window{}
	w.Styles = element.Styles{
		PsuedoStyles: map[string]map[string]map[string]string{},
		StyleMap:     map[string][]*element.StyleMap{},
	}
	css := cstyle.CSS{
		Width:   float32(width),
		Height:  float32(height),
		Adapter: adapterFunction,
	}

	w.Styles.StyleTag(mastercss)
	// This is still apart of computestyle
	css.AddPlugin(inline.Init())
	css.AddPlugin(textAlign.Init())
	css.AddPlugin(flex.Init())
	css.AddPlugin(crop.Init())

	css.AddTransformer(text.Init())
	css.AddTransformer(banda.Init())
	css.AddTransformer(scrollbar.Init())
	css.AddTransformer(marginblock.Init())
	css.AddTransformer(ul.Init())
	css.AddTransformer(ol.Init())

	el := element.Node{}
	document := el.CreateElement("ROOT")
	document.Properties.Id = "ROOT"
	document.StyleSheets = &w.Styles
	s := scripts.Scripts{}
	s.Add(a.Init())

	w.CSS = css
	w.Scripts = s
	w.document = document

	return w
}

func (w *Window) Open() {
	for !w.shouldStop {
		w.CSS.Adapter.Render(w.RenderData)

	}
}

func flatten(n *element.Node) []*element.Node {
	var nodes []*element.Node
	nodes = append(nodes, n)

	children := n.Children
	if len(children) > 0 {
		for _, ch := range children {
			chNodes := flatten(ch)
			nodes = append(nodes, chNodes...)
		}
	}
	return nodes
}

func open(data *Window) {
	data.document.ComputedStyle["width"] = strconv.Itoa(int(data.CSS.Width)) + "px"
	data.document.ComputedStyle["height"] = strconv.Itoa(int(data.CSS.Height)) + "px"

	data.CSS.Adapter.Init(int(data.CSS.Width), int(data.CSS.Height))

	data.CSS.State = map[string]element.State{}
	data.CSS.State["ROOT"] = element.State{
		Width:  float32(data.CSS.Width),
		Height: float32(data.CSS.Height),
	}

	// Load init font
	if data.CSS.Fonts == nil {
		data.CSS.Fonts = map[string]*truetype.Font{}
	}
	fid := "Georgia 16px false false"
	if data.CSS.Fonts[fid] == nil {
		f, _ := font.LoadFont("Georgia", 16, "", false, &data.CSS.Adapter.FileSystem)
		data.CSS.Fonts[fid] = f
	}

	monitor := events.Monitor{
		EventMap: make(map[string]element.Event),
		Adapter:  data.CSS.Adapter,
		CSS:      &data.CSS,
		Focus: events.Focus{
			Nodes:               []string{},
			Selected:            -1,
			SoftFocused:         "",
			LastClickWasFocused: false,
		},
	}

	data.CSS.Adapter.AddEventListener("windowresize", func(e element.Event) {
		wh := e.Data.(map[string]int)

		data.CSS.Width = float32(wh["width"])
		data.CSS.Height = float32(wh["height"])

		data.document.ComputedStyle["width"] = strconv.Itoa(wh["width"]) + "px"
		data.document.ComputedStyle["height"] = strconv.Itoa(wh["height"]) + "px"
		getRenderData(data, &monitor)
	})

	data.CSS.Adapter.AddEventListener("close", func(e element.Event) {
		data.shouldStop = true
	})

	currentEvent := events.EventData{}

	data.CSS.Adapter.AddEventListener("keydown", func(e element.Event) {
		currentEvent.Key = e.Data.(int)
		currentEvent.KeyState = true
		currentEvent.Modifiers = events.Modifiers{
			CtrlKey:  e.CtrlKey,
			ShiftKey: e.ShiftKey,
			MetaKey:  e.MetaKey,
			AltKey:   e.AltKey,
		}
		monitor.GetEvents(&currentEvent)
		getRenderData(data, &monitor)
	})
	data.CSS.Adapter.AddEventListener("keyup", func(e element.Event) {
		currentEvent.Key = 0
		currentEvent.KeyState = false
		currentEvent.Modifiers = events.Modifiers{
			CtrlKey:  e.CtrlKey,
			ShiftKey: e.ShiftKey,
			MetaKey:  e.MetaKey,
			AltKey:   e.AltKey,
		}
		monitor.GetEvents(&currentEvent)
		getRenderData(data, &monitor)
	})

	data.CSS.Adapter.AddEventListener("mousemove", func(e element.Event) {
		pos := e.Data.([]int)
		if pos[0] > 0 && pos[1] > 0 {
			if pos[0] < int(data.CSS.Width) && pos[1] < int(data.CSS.Height) {
				currentEvent.Position = pos
				monitor.GetEvents(&currentEvent)
				getRenderData(data, &monitor)
			}
		}
	})

	data.CSS.Adapter.AddEventListener("scroll", func(e element.Event) {
		currentEvent.ScrollY = e.Data.(int)
		monitor.GetEvents(&currentEvent)
		currentEvent.ScrollY = 0
		getRenderData(data, &monitor)
	})

	data.CSS.Adapter.AddEventListener("mousedown", func(e element.Event) {
		currentEvent.Click = true
		monitor.GetEvents(&currentEvent)
		getRenderData(data, &monitor)
	})

	data.CSS.Adapter.AddEventListener("mouseup", func(e element.Event) {
		currentEvent.Click = false
		monitor.GetEvents(&currentEvent)
		getRenderData(data, &monitor)
	})

	data.CSS.Adapter.AddEventListener("contextmenudown", func(e element.Event) {
		currentEvent.Context = true
		monitor.GetEvents(&currentEvent)
		getRenderData(data, &monitor)
	})

	data.CSS.Adapter.AddEventListener("contextmenuup", func(e element.Event) {
		currentEvent.Context = true
		monitor.GetEvents(&currentEvent)
		getRenderData(data, &monitor)
	})

	getRenderData(data, &monitor)

}

// !TODO: This need to be better implemented but rn just testing
func getRenderData(data *Window, monitor *events.Monitor) {
	data.CSS.State["ROOT"] = element.State{
		Width:  float32(data.CSS.Width),
		Height: float32(data.CSS.Height),
	}
	fmt.Println("_______________________")

	monitor.RunEvents(data.document.Children[0])
	start := time.Now()
	newDoc := copyDocument(data.CSS, data.document.Children[0], &data.document)

	data.CSS.ComputeNodeStyle(newDoc)

	flatDoc := flatten(newDoc)

	rd := []element.State{}

	keys := []string{}
	s := data.CSS.State
	for _, v := range flatDoc {
		rd = append(rd, s[v.Properties.Id])
		keys = append(keys, v.Properties.Id)
	}

	// Create a set of keys to keep
	keysSet := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		keysSet[key] = struct{}{}
	}

	for k, self := range s {
		key := cstyle.BackgroundKey(self)
		if _, found := keysSet[k]; !found {
			for t := range data.CSS.Adapter.Textures[k] {
				data.CSS.Adapter.UnloadTexture(k, t)
			}
			delete(s, k)
		} else {
			if data.CSS.Adapter.Textures[k]["background"] != key {
				img := cstyle.GenerateBackground(data.CSS, self)
				data.CSS.Adapter.LoadTexture(k, "background", key, img)
				if self.Textures == nil {
					self.Textures = map[string]string{}
				}

				self.Textures["background"] = key
				data.CSS.State[k] = self
			}
		}
	}

	addScroll(&data.document, s)

	data.Scripts.Run(&data.document)

	// !TODO: Should return effected node, then render those specific
	// + I think have node.ComputeNodeStyle would make this nice

	fmt.Println(time.Since(start))
	data.RenderData = rd
	(data.CSS.State) = s
}

func copyDocument(c cstyle.CSS, node *element.Node, parent *element.Node) *element.Node {
	n := *node
	n.Parent = parent
	// !DEVMAN: Copying is done here, would like to remove this and add it to ComputeNodeStyle, so I can save a tree climb
	// + Maybe just have this copy the node and the styles don't need to be recomputed everytime

	if len(node.Children) > 0 {
		n.Children = make([]*element.Node, 0, len(node.Children))
		// n.Children = append(n.Children, node.Children...)
		// for _, v := range node.Children {
		for i := range node.Children {
			n.Children = append(n.Children, node.Children[i])
		}
		for i := range node.Children {
			n.Children[i] = copyDocument(c, node.Children[i], &n)
		}

	}

	return &n
}

func addScroll(n *element.Node, s map[string]element.State) {
	// !NOTE: This is the only spot you can pierce the vale
	n.ScrollHeight = s[n.Properties.Id].ScrollHeight
	n.ScrollWidth = s[n.Properties.Id].ScrollWidth
	for i := range n.Children {
		addScroll(n.Children[i], s)
	}
}

func createNode(node *html.Node, parent *element.Node, stylesheets *element.Styles) {
	if node.Type == html.ElementNode {
		newNode := parent.CreateElement(node.Data)
		newNode.Parent = parent
		newNode.Properties.Id = element.GenerateUniqueId(parent, node.Data)
		for _, attr := range node.Attr {
			switch attr.Key {
			case "class":
				classes := strings.Split(attr.Val, " ")
				for _, class := range classes {
					newNode.ClassList.Add(class)
				}
			case "id":
				newNode.Id = attr.Val
			case "contenteditable":
				if attr.Val == "" || attr.Val == "true" {
					newNode.ContentEditable = true
				}
			case "href":
				newNode.Href = attr.Val
			case "src":
				newNode.Src = attr.Val
			case "title":
				newNode.Title = attr.Val
			case "tabindex":
				newNode.TabIndex, _ = strconv.Atoi(attr.Val)
			case "disabled":
				newNode.Disabled = true
			case "required":
				newNode.Required = true
			case "checked":
				newNode.Checked = true
			default:
				newNode.SetAttribute(attr.Key, attr.Val)
			}
		}
		newNode.InnerText = strings.TrimSpace(utils.GetInnerText(node))
		parent.StyleSheets.GetStyles(&newNode)
		// Recursively traverse child nodes
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			if child.Type == html.ElementNode {
				createNode(child, &newNode, stylesheets)
			}
		}
		parent.AppendChild(&newNode)
	} else {
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			if child.Type == html.ElementNode {
				createNode(child, parent, stylesheets)
			}
		}
	}
}

func parseHTMLFromFile(path string, fs adapter.FileSystem) ([]string, []string, *html.Node) {
	file, _ := fs.ReadFile(path)

	doc, _ := html.Parse(strings.NewReader(string(file)))
	wrapAllTextNodes(doc)
	unwrapSingleTextChildren(doc)

	// Extract stylesheet link tags and style tags
	stylesheets := extractStylesheets(doc, filepath.Dir(path))
	styleTags := extractStyleTags(doc)

	return stylesheets, styleTags, doc
}

func extractStylesheets(n *html.Node, baseDir string) []string {
	var stylesheets []string

	var dfs func(*html.Node)
	dfs = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "link" {
			var href string
			isStylesheet := false

			for _, attr := range node.Attr {
				if attr.Key == "rel" && attr.Val == "stylesheet" {
					isStylesheet = true
				} else if attr.Key == "href" {
					href = attr.Val
				}
			}

			if isStylesheet {
				resolvedHref := localizePath(baseDir, href)
				stylesheets = append(stylesheets, resolvedHref)
			}
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			dfs(c)
		}
	}

	dfs(n)
	return stylesheets
}

func extractStyleTags(n *html.Node) []string {
	var styleTags []string

	var dfs func(*html.Node)
	dfs = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "style" {
			var styleContent strings.Builder
			for c := node.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.TextNode {
					styleContent.WriteString(c.Data)
				}
			}
			styleTags = append(styleTags, styleContent.String())
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			dfs(c)
		}
	}

	dfs(n)
	return styleTags
}

func localizePath(rootPath, filePath string) string {
	// Check if the file path has a scheme, indicating it's a URL
	u, err := url.Parse(filePath)
	if err == nil && u.Scheme != "" {
		return filePath
	}

	// Join the root path and the file path to create an absolute path
	absPath := filepath.Join(rootPath, filePath)

	// If the absolute path is the same as the original path, return it
	if absPath == filePath {
		return filePath
	}

	return "./" + absPath
}

// wrapAllTextNodes wraps all non-empty text nodes with <text> elements
func wrapAllTextNodes(n *html.Node) {
	// Skip script and style tags
	if n.Type == html.ElementNode && (n.Data == "script" || n.Data == "style") {
		return
	}

	// Process children (collect first to avoid traversal issues)
	var children []*html.Node
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		children = append(children, c)
	}

	for _, child := range children {
		// Wrap text nodes
		child.Data = strings.TrimSpace(child.Data)
		child.Data = strings.ReplaceAll(child.Data, "\t", " ")

		// Replace repeating spaces with a single space
		for strings.Contains(child.Data, "  ") {
			child.Data = strings.ReplaceAll(child.Data, "  ", " ")
		}

		if child.Type == html.TextNode && child.Data != "" {
			textEl := &html.Node{
				Type: html.ElementNode,
				Data: "text",
			}

			n.InsertBefore(textEl, child)
			n.RemoveChild(child)
			textEl.AppendChild(child)
		}

		// Process child's children recursively
		wrapAllTextNodes(child)
	}
}

// unwrapSingleTextChildren removes <text> elements that are the only child of their parent
func unwrapSingleTextChildren(n *html.Node) {
	// Skip script and style tags
	if n.Type == html.ElementNode && (n.Data == "script" || n.Data == "style") {
		return
	}

	// Check if this element has exactly one child and it's a <text> element
	if n.Type == html.ElementNode && n.Data != "text" && n.FirstChild != nil &&
		n.FirstChild.NextSibling == nil && n.FirstChild.Type == html.ElementNode &&
		n.FirstChild.Data == "text" {

		textEl := n.FirstChild
		textContent := textEl.FirstChild

		if textContent != nil {
			// Move the text content directly under this element
			textEl.RemoveChild(textContent)
			n.InsertBefore(textContent, textEl)
			n.RemoveChild(textEl)
		}
	}

	// Process all children recursively
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		unwrapSingleTextChildren(c)
	}
}
