package document

import (
	"bufio"
	"gui/cstyle"
	"gui/window"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/go-shiori/dom"
	"golang.org/x/net/html"
)

type Doc struct {
	StyleSheets []string
	StyleTags   []string
	DOM         *html.Node
	Title       string
}

func Open(index string) {
	d := Parse(index)

	wm := window.NewWindowManager()
	wm.FPS = true

	// Initialization
	var screenWidth int32 = 800
	var screenHeight int32 = 450

	// Open the window
	wm.OpenWindow(d.Title, screenWidth, screenHeight)
	defer wm.CloseWindow()

	css := cstyle.CSS{
		Width:  800,
		Height: 450,
	}
	css.StyleSheet("./master.css")

	for _, v := range d.StyleSheets {
		css.StyleSheet(v)
	}

	for _, v := range d.StyleTags {
		css.StyleTag(v)
	}

	nodes := css.Map(d.DOM)
	wm.LoadTextures(nodes.Render)
	wm.Draw(nodes.Render)

	// Main game loop
	for !wm.WindowShouldClose() {
		rl.BeginDrawing()
		// rl.ClearBackground(rl.RayWhite)
		// Check if the window size has changed
		newWidth := int32(rl.GetScreenWidth())
		newHeight := int32(rl.GetScreenHeight())

		if newWidth != screenWidth || newHeight != screenHeight {
			rl.ClearBackground(rl.RayWhite)
			// Window has been resized, handle the event
			screenWidth = newWidth
			screenHeight = newHeight

			css.Width = float32(screenWidth)
			css.Height = float32(screenHeight)
			nodes = css.Map(d.DOM)
			wm.LoadTextures(nodes.Render)
		}
		wm.Draw(nodes.Render)

		rl.EndDrawing()
	}
}

func Parse(path string) Doc {
	file, err := os.Open(path)
	check(err)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var htmlContent string

	for scanner.Scan() {
		htmlContent += scanner.Text() + "\n"
	}

	check(scanner.Err())

	doc, err := dom.FastParse(strings.NewReader(encapsulateText(htmlContent)))
	check(err)

	// Extract stylesheet link tags and style tags
	stylesheets := extractStylesheets(doc, filepath.Dir(path))
	styleTags := extractStyleTags(doc)

	d := Doc{
		StyleSheets: stylesheets,
		StyleTags:   styleTags,
		DOM:         doc,
		Title:       dom.InnerText(dom.GetElementsByTagName(doc, "title")[0]),
	}

	return d
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
	htmlString = removeHTMLComments(htmlString)
	openClose := regexp.MustCompile(`(?s)(<\w+(?:\s+\w+="[^"]*")*>)([^<]+)(<\w+\s*(?:\s+\w+="[^"]*")*>)`)
	closeOpen := regexp.MustCompile(`(?s)(</\w+(?:\s+\w+="[^"]*")*>)([^<]+)(<\w+\s*(?:\s+\w+="[^"]*")*>)`)
	closeClose := regexp.MustCompile(`(?s)(</\w+(?:\s+\w+="[^"]*")*>)([^<]+)(</\w+\s*(?:\s+\w+="[^"]*")*>)`)
	a := matchFactory(openClose)
	t := openClose.ReplaceAllStringFunc(htmlString, a)
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
			return submatches[1] + "<span>" + submatches[2] + "</span>" + submatches[3]
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
	re := regexp.MustCompile(`<!--(.*?)-->`)
	return re.ReplaceAllString(htmlString, "")
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}