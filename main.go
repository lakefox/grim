package grim

import (
	_ "embed"
	"fmt"
	adapter "grim/adapters"
	"grim/canvas"
	"grim/cstyle"
	"grim/cstyle/plugins/crop"
	"grim/cstyle/plugins/flex"
	"grim/cstyle/plugins/inline"
	"grim/cstyle/plugins/textAlign"
	"grim/cstyle/transformers/background"
	"grim/cstyle/transformers/banda"
	flexprep "grim/cstyle/transformers/flex"
	img "grim/cstyle/transformers/image"
	"time"

	marginblock "grim/cstyle/transformers/margin-block"
	"grim/cstyle/transformers/ol"
	"grim/cstyle/transformers/scrollbar"
	// "grim/cstyle/transformers/text"
	"grim/cstyle/transformers/ul"
	"grim/font"
	"grim/library"
	"grim/scripts"
	"grim/scripts/a"
	"image"

	"grim/element"
	"grim/events"
	"grim/utils"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	imgFont "golang.org/x/image/font"

	"golang.org/x/net/html"
)

//go:embed master.css
var mastercss string

type Window struct {
	CSS cstyle.CSS
	// document element.Node

	Scripts scripts.Scripts
}

func (w *Window) Document() *element.Node {
	return w.CSS.Document["ROOT"].Children[0]
}

// !TODO: Add a Mux option to all a http server to map to the window
func (window *Window) HttpMux() {}

func (window *Window) Path(path string) {
	styleSheets, styleTags, htmlNodes := parseHTMLFromFile(path, window.CSS.Adapter.FileSystem)

	for _, v := range styleSheets {
		window.CSS.StyleSheet(v)
	}

	for _, v := range styleTags {
		window.CSS.StyleTag(v)
	}

	window.CSS.Path = filepath.Dir(path)

	CreateNode(htmlNodes, window.CSS.Document["ROOT"], &window.CSS)
	open(window)
}

func New(adapterFunction *adapter.Adapter, width, height int) Window {
	css := cstyle.CSS{
		Width:   float32(width),
		Height:  float32(height),
		Adapter: adapterFunction,
	}

	css.StyleTag(mastercss)
	// This is still apart of computestyle
	css.AddPlugin(inline.Init())
	css.AddPlugin(textAlign.Init())
	css.AddPlugin(flex.Init())
	css.AddPlugin(crop.Init())

	// css.AddTransformer(text.Init())
	css.AddTransformer(banda.Init())
	css.AddTransformer(scrollbar.Init())
	css.AddTransformer(flexprep.Init())
	css.AddTransformer(marginblock.Init())
	css.AddTransformer(ul.Init())
	css.AddTransformer(ol.Init())
	css.AddTransformer(background.Init())
	css.AddTransformer(img.Init())

	el := element.Node{}
	document := el.CreateElement("ROOT")
	// document.CStyle["width"] = strconv.Itoa(width) + "px"
	// document.CStyle["height"] = strconv.Itoa(height) + "px"
	document.Properties.Id = "ROOT"

	s := scripts.Scripts{}
	s.Add(a.Init())

	css.Document = map[string]*element.Node{
		"ROOT": &document,
	}

	return Window{
		CSS:     css,
		Scripts: s,
	}
}

func (w *Window) Render(doc *element.Node, shelf *library.Shelf) []element.State {
	s := w.CSS.State

	flatDoc := flatten(doc)

	store := []element.State{}

	keys := []string{}

	for _, v := range flatDoc {
		store = append(store, s[v.Properties.Id])
		keys = append(keys, v.Properties.Id)
	}

	// Create a set of keys to keep
	keysSet := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		keysSet[key] = struct{}{}
	}

	// Iterate over the map and delete keys not in the set
	for k := range s {
		if _, found := keysSet[k]; !found {
			delete(s, k)
		}
	}

	for k, self := range store {
		// Option: Have Grim render all elements
		wbw := int(self.Width + self.Border.Left.Width + self.Border.Right.Width)
		hbw := int(self.Height + self.Border.Top.Width + self.Border.Bottom.Width)

		key := strconv.Itoa(wbw) + strconv.Itoa(hbw) + utils.RGBAtoString(self.Background)

		exists := shelf.Check(key)
		bounds := shelf.Bounds(key)
		// fmt.Println(n.Properties.Id, self.Width, self.Height, bounds)

		if exists && bounds[0] == int(wbw) && bounds[1] == int(hbw) {
			lookup := make(map[string]struct{}, len(self.Textures))
			for _, v := range self.Textures {
				lookup[v] = struct{}{}
			}

			if _, found := lookup[key]; !found {
				self.Textures = append([]string{key}, self.Textures...)
				store[k] = self
			}
		} else if self.Background.A > 0 {
			lookup := make(map[string]struct{}, len(self.Textures))
			for _, v := range self.Textures {
				lookup[v] = struct{}{}
			}

			if _, found := lookup[key]; !found {
				// Only make the drawing if it's not found
				can := canvas.NewCanvas(wbw, hbw)
				can.BeginPath()
				can.SetFillStyle(self.Background.R, self.Background.G, self.Background.B, self.Background.A)
				can.SetLineWidth(10)
				can.RoundedRect(0, 0, float64(wbw), float64(hbw),
					[]float64{float64(self.Border.Radius.TopLeft), float64(self.Border.Radius.TopRight), float64(self.Border.Radius.BottomRight), float64(self.Border.Radius.BottomLeft)})
				can.Fill()
				can.ClosePath()

				shelf.Set(key, can.RGBA)
				self.Textures = append([]string{key}, self.Textures...)
				store[k] = self
			}
		}
	}

	return store
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

// !ISSUE: Probally don't need this to be exposed to the outside, if using getter/setter, just render once content is loaded then everytime a event
// + or content update
func open(data *Window) {
	shelf := library.Shelf{
		Textures:   map[string]*image.RGBA{},
		References: map[string]bool{},
	}

	debug := false
	data.CSS.Styles = map[string]map[string]string{}
	data.CSS.PsuedoStyles = map[string]map[string]map[string]string{}

	data.CSS.Styles["ROOT"] = map[string]string{}
	data.CSS.Styles["ROOT"]["width"] = strconv.Itoa(int(data.CSS.Width)) + "px"
	data.CSS.Styles["ROOT"]["height"] = strconv.Itoa(int(data.CSS.Height)) + "px"

	data.CSS.Adapter.Library = &shelf
	data.CSS.Adapter.Init(int(data.CSS.Width), int(data.CSS.Height))

	data.CSS.State = map[string]element.State{}
	data.CSS.State["ROOT"] = element.State{
		Width:  float32(data.CSS.Width),
		Height: float32(data.CSS.Height),
	}

	shouldStop := false

	var rd []element.State

	// Load init font
	if data.CSS.Fonts == nil {
		data.CSS.Fonts = map[string]imgFont.Face{}
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

		data.CSS.Styles["ROOT"]["width"] = strconv.Itoa(wh["width"]) + "px"
		data.CSS.Styles["ROOT"]["height"] = strconv.Itoa(wh["height"]) + "px"
		rd = getRenderData(data, &shelf, &monitor)
	})

	data.CSS.Adapter.AddEventListener("close", func(e element.Event) {
		shouldStop = true
	})

	currentEvent := events.EventData{}

	data.CSS.Adapter.AddEventListener("keydown", func(e element.Event) {
		currentEvent.Key = e.Data.(int)
		currentEvent.KeyState = true
		monitor.GetEvents(&currentEvent)
	})
	data.CSS.Adapter.AddEventListener("keyup", func(e element.Event) {
		currentEvent.Key = 0
		currentEvent.KeyState = false
		monitor.GetEvents(&currentEvent)
	})

	data.CSS.Adapter.AddEventListener("mousemove", func(e element.Event) {
		pos := e.Data.([]int)
		if pos[0] > 0 && pos[1] > 0 {
			if pos[0] < int(data.CSS.Width) && pos[1] < int(data.CSS.Height) {
				currentEvent.Position = pos
				monitor.GetEvents(&currentEvent)
				rd = getRenderData(data, &shelf, &monitor)
			}
		}
	})

	data.CSS.Adapter.AddEventListener("scroll", func(e element.Event) {
		currentEvent.ScrollY = e.Data.(int)
		monitor.GetEvents(&currentEvent)
		currentEvent.ScrollY = 0
		rd = getRenderData(data, &shelf, &monitor)
	})

	data.CSS.Adapter.AddEventListener("mousedown", func(e element.Event) {
		currentEvent.Click = true
		monitor.GetEvents(&currentEvent)
		rd = getRenderData(data, &shelf, &monitor)
	})

	data.CSS.Adapter.AddEventListener("mouseup", func(e element.Event) {
		currentEvent.Click = false
		monitor.GetEvents(&currentEvent)
		rd = getRenderData(data, &shelf, &monitor)
	})

	data.CSS.Adapter.AddEventListener("contextmenudown", func(e element.Event) {
		currentEvent.Context = true
		monitor.GetEvents(&currentEvent)
	})

	data.CSS.Adapter.AddEventListener("contextmenuup", func(e element.Event) {
		currentEvent.Context = true
		monitor.GetEvents(&currentEvent)
	})

	// !ISSUE: the loop should be moved to the adapter and the rerendering should only happen if a eventlistener goes off
	// + also have a animation loop seperate but thats later
	// + really should run runevents after get events and if there is any changes then rerender
	// + ahh what about dom changes in the js api...
	// + could swap to getters and setters but i don't like them
	// Main game loop
	rd = getRenderData(data, &shelf, &monitor)
	// !TODO: Move to adapter

	for !shouldStop {
		if !shouldStop && debug {
			shouldStop = true
		}
		// Check if the window size has changed

		data.CSS.Adapter.Render(rd)
	}
}

// !TODO: This need to be better implemented but rn just testing
func getRenderData(data *Window, shelf *library.Shelf, monitor *events.Monitor) []element.State {
	data.CSS.State["ROOT"] = element.State{
		Width:  float32(data.CSS.Width),
		Height: float32(data.CSS.Height),
	}
	fmt.Println("_______________________")

	// !ISSUE: Move this inside of ComputeNodeStyle
	// + This modify events to return effected nodes

	dc := data.CSS.Document["ROOT"].Children[0]
	start := time.Now()
	newDoc := AddStyles(data.CSS, dc, data.CSS.Document["ROOT"])

	data.CSS.ComputeNodeStyle(newDoc)
	fmt.Println(time.Since(start))

	rd := data.Render(newDoc, shelf)

	data.CSS.Adapter.Load(rd)

	AddHTMLAndAttrs(data.CSS.Document["ROOT"], data.CSS.State)
	// fmt.Println(data.document.InnerHTML)

	data.Scripts.Run(data.CSS.Document["ROOT"])
	shelf.Clean()

	// !TODO: Should return effected node, then render those specific
	// + I think have node.ComputeNodeStyle would make this nice
	monitor.RunEvents(data.CSS.Document["ROOT"].Children[0])
	return rd
}

func AddStyles(c cstyle.CSS, node *element.Node, parent *element.Node) *element.Node {
	n := *node
	n.Parent = parent
	// !DEVMAN: Copying is done here, would like to remove this and add it to ComputeNodeStyle, so I can save a tree climb
	c.GetStyles(&n)

	if len(node.Children) > 0 {
		n.Children = make([]*element.Node, 0, len(node.Children))
		// n.Children = append(n.Children, node.Children...)
		// for _, v := range node.Children {
		for i := 0; i < len(node.Children); i++ {
			n.Children = append(n.Children, node.Children[i])
		}
		for i := 0; i < len(node.Children); i++ {
			n.Children[i] = AddStyles(c, node.Children[i], &n)
		}

	}

	return &n
}

func CreateNode(node *html.Node, parent *element.Node, css *cstyle.CSS) {
	if node.Type == html.ElementNode {
		newNode := parent.CreateElement(node.Data)
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
		
		newNode.Properties.Id = element.GenerateUniqueId(parent, node.Data)
		// Recursively traverse child nodes
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			if child.Type == html.ElementNode {
				CreateNode(child, &newNode, css)
			}
		}
		parent.AppendChild(&newNode)
		css.Document[newNode.Properties.Id] = &newNode

	} else {
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			if child.Type == html.ElementNode {
				CreateNode(child, parent, css)
			}
		}
	}
}

func AddHTMLAndAttrs(n *element.Node, s map[string]element.State) {
	// Head is not renderable

	n.InnerHTML = utils.InnerHTML(n)
	tag, closing := utils.NodeToHTML(n)
	n.OuterHTML = tag + n.InnerHTML + closing

	// !NOTE: This is the only spot you can pierce the vale
	n.ScrollHeight = s[n.Properties.Id].ScrollHeight
	n.ScrollWidth = s[n.Properties.Id].ScrollWidth
	for i := range n.Children {
		AddHTMLAndAttrs(n.Children[i], s)
	}
}

func parseHTMLFromFile(path string, fs adapter.FileSystem) ([]string, []string, *html.Node) {
	file, _ := fs.ReadFile(path)

	htmlContent := removeHTMLComments(string(file))

	doc, _ := html.Parse(strings.NewReader(encapsulateText(removeWhitespaceBetweenTags(htmlContent))))

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

func encapsulateText(htmlString string) string {
	openOpen := regexp.MustCompile(`(<\w+[^>]*>)([^<]+)(<\w+[^>]*>)`)
	closeOpen := regexp.MustCompile(`(</\w+[^>]*>)([^<]+)(<\w+[^>]*>)`)
	closeClose := regexp.MustCompile(`(<\/\w+[^>]*>)([^<]+)(<\/\w+[^>]*>)`)
	a := matchFactory(openOpen)
	t := openOpen.ReplaceAllStringFunc(htmlString, a)
	b := matchFactory(closeOpen)
	u := closeOpen.ReplaceAllStringFunc(t, b)
	c := matchFactory(closeClose)
	v := closeClose.ReplaceAllStringFunc(u, c)
	return v
}

func matchFactory(re *regexp.Regexp) func(string) string {
	return func(match string) string {
		submatches := re.FindStringSubmatch(match)
		if len(submatches) != 4 {
			return match
		}

		// Process submatches
		if len(removeWhitespace(submatches[2])) > 0 {
			return submatches[1] + "<text>" + submatches[2] + "</text>" + submatches[3]
		} else {
			return match
		}
	}
}
func removeWhitespace(htmlString string) string {
	// Remove extra white space
	reSpaces := regexp.MustCompile(`\s+`)
	htmlString = reSpaces.ReplaceAllString(htmlString, " ")

	// Trim leading and trailing white space
	htmlString = strings.TrimSpace(htmlString)

	return htmlString
}

func removeHTMLComments(htmlString string) string {
	re := regexp.MustCompile(`<!--[\s\S]*?-->`)
	return re.ReplaceAllString(htmlString, "")
}

// important to allow the notspans to be injected, the spaces after removing the comments cause the regexp to fail
func removeWhitespaceBetweenTags(html string) string {
	// Create a regular expression to match spaces between angle brackets
	re := regexp.MustCompile(`>\s+<`)
	// Replace all matches of spaces between angle brackets with "><"
	return re.ReplaceAllString(html, "><")
}
