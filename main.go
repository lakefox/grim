package grim

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/json"
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
	marginblock "grim/cstyle/transformers/margin-block"
	"grim/cstyle/transformers/ol"
	"grim/cstyle/transformers/scrollbar"
	"grim/cstyle/transformers/text"
	"grim/cstyle/transformers/ul"
	"grim/font"
	"grim/library"
	"grim/scripts"
	"grim/scripts/a"
	"image"
	"time"

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
	CSS      cstyle.CSS
	document element.Node

	Scripts scripts.Scripts
}

func (w *Window) Document() *element.Node {
	return w.document.Children[0]
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

	CreateNode(htmlNodes, &window.document)
	// window.Document = *window.Document
}

func New(adapterFunction *adapter.Adapter) Window {
	css := cstyle.CSS{
		Width:   800,
		Height:  450,
		Adapter: adapterFunction,
	}

	css.StyleTag(mastercss)
	// This is still apart of computestyle
	css.AddPlugin(inline.Init())
	css.AddPlugin(textAlign.Init())
	css.AddPlugin(flex.Init())
	css.AddPlugin(crop.Init())

	css.AddTransformer(text.Init())
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
	document.Style["width"] = "800px"
	document.Style["height"] = "450px"
	document.Properties.Id = "ROOT"

	s := scripts.Scripts{}
	s.Add(a.Init())

	return Window{
		CSS:      css,
		document: document,
		Scripts:  s,
	}
}

func (w *Window) Render(doc *element.Node, state *map[string]element.State, shelf *library.Shelf) []element.State {
	s := *state

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
func Open(data *Window, width, height int) {
	shelf := library.Shelf{
		Textures:   map[string]*image.RGBA{},
		References: map[string]bool{},
	}

	debug := false
	data.document.Style["width"] = strconv.Itoa(int(width)) + "px"
	data.document.Style["height"] = strconv.Itoa(int(height)) + "px"

	data.CSS.Adapter.Library = &shelf
	data.CSS.Adapter.Init(width, height)

	state := map[string]element.State{}
	state["ROOT"] = element.State{
		Width:  float32(width),
		Height: float32(height),
	}

	shouldStop := false

	var hash []byte
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

	newWidth, newHeight := width, height

	monitor := events.Monitor{
		EventMap: make(map[string]element.Event),
		Adapter:  data.CSS.Adapter,
		State:    &state,
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
		newWidth = wh["width"]
		newHeight = wh["height"]
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
		currentEvent.Position = pos
		monitor.GetEvents(&currentEvent)
	})

	data.CSS.Adapter.AddEventListener("scroll", func(e element.Event) {
		currentEvent.ScrollY = e.Data.(int)
		monitor.GetEvents(&currentEvent)
		currentEvent.ScrollY = 0
	})

	data.CSS.Adapter.AddEventListener("mousedown", func(e element.Event) {
		currentEvent.Click = true
		monitor.GetEvents(&currentEvent)
	})

	data.CSS.Adapter.AddEventListener("mouseup", func(e element.Event) {
		currentEvent.Click = false
		monitor.GetEvents(&currentEvent)
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
	for !shouldStop {

		if !shouldStop && debug {
			shouldStop = true
		}
		// Check if the window size has changed
		resize := false

		if newWidth != width || newHeight != height {
			resize = true
			// Window has been resized, handle the event
			width = newWidth
			height = newHeight

			data.CSS.Width = float32(width)
			data.CSS.Height = float32(height)

			data.document.Style["width"] = strconv.Itoa(int(width)) + "px"
			data.document.Style["height"] = strconv.Itoa(int(height)) + "px"
		}

		newHash, _ := hashStruct(&data.document.Children[0])

		if !bytes.Equal(hash, newHash) || resize {
			hash = newHash
			fmt.Println("----------------------------------")
			s := time.Now()
			newDoc := AddStyles(data.CSS, data.document.Children[0], &data.document)

			newDoc = data.CSS.Transform(newDoc)

			state["ROOT"] = element.State{
				Width:  float32(width),
				Height: float32(height),
			}

			data.CSS.ComputeNodeStyle(newDoc, &state)

			rd = data.Render(newDoc, &state, &shelf)

			data.CSS.Adapter.Load(rd)

			AddHTMLAndAttrs(&data.document, &state)
			// AddHTMLAndAttrs(newDoc, &state)
			// fmt.Println(newDoc.OuterHTML)
			data.Scripts.Run(&data.document)
			shelf.Clean()
			elapsed := time.Since(s)
			fmt.Printf("Execution time: %s\n", elapsed)
		}

		monitor.RunEvents(data.document.Children[0])
		data.CSS.Adapter.Render(rd)
	}
}

func AddStyles(c cstyle.CSS, node *element.Node, parent *element.Node) *element.Node {
	n := *node
	n.Parent = parent

	n.Style, n.PseudoElements = c.GetStyles(&n)

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

func CreateNode(node *html.Node, parent *element.Node) {
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
		// Recursively traverse child nodes
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			if child.Type == html.ElementNode {
				CreateNode(child, &newNode)
			}
		}
		parent.AppendChild(&newNode)

	} else {
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			if child.Type == html.ElementNode {
				CreateNode(child, parent)
			}
		}
	}
}

func AddHTMLAndAttrs(n *element.Node, state *map[string]element.State) {
	// Head is not renderable
	s := (*state)
	n.InnerHTML = utils.InnerHTML(n)
	tag, closing := utils.NodeToHTML(n)
	n.OuterHTML = tag + n.InnerHTML + closing
	// !NOTE: This is the only spot you can pierce the vale
	n.ScrollHeight = s[n.Properties.Id].ScrollHeight
	n.ScrollWidth = s[n.Properties.Id].ScrollWidth
	for i := range n.Children {
		AddHTMLAndAttrs(n.Children[i], state)
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
			return submatches[1] + "<notaspan>" + submatches[2] + "</notaspan>" + submatches[3]
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

// Function to hash a struct using SHA-256
func hashStruct(s interface{}) ([]byte, error) {
	// Convert struct to JSON
	jsonData, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}

	// Hash the JSON data using SHA-256
	hasher := sha256.New()
	hasher.Write(jsonData)
	hash := hasher.Sum(nil)

	return hash, nil
}
