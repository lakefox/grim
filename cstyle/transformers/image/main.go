package img

import (
	"bytes"
	"grim/canvas"
	"grim/cstyle"
	"grim/element"
	"image"
	_ "image/gif"  // Enable GIF support
	_ "image/jpeg" // Enable JPEG support
	_ "image/png"  // Enable PNG support
	"path/filepath"
	"strconv"
)

func Init() cstyle.Transformer {
	return cstyle.Transformer{
		Selector: func(n *element.Node, c *cstyle.CSS) bool {
			return n.TagName == "img"
			// !ISSUE: img tags or background-url || background: url()
		},
		Handler: func(n *element.Node, c *cstyle.CSS) *element.Node {
			n.TagName = "canvas"
			if !c.Adapter.Library.Check(n.Properties.Id + "canvas") {
				i, err := c.Adapter.FileSystem.ReadFile(filepath.Join(c.Path, n.Src))
				if err == nil {
					img, _, err := image.Decode(bytes.NewReader(i))
					if err == nil {
						width, height := n.GetStyle("width"), n.GetStyle("height")
						if n.GetStyle("width") == "" {
							width = strconv.Itoa(img.Bounds().Dx()) + "px"
						}
						if n.GetStyle("height") == "" {
							height = strconv.Itoa(img.Bounds().Dy()) + "px"
						}
						ctx := n.GetContext(img.Bounds().Dx(), img.Bounds().Dy())
						n.SetStyle("width", width)
						n.SetStyle("height", height)
						ctx.DrawImage(img, 0, 0)
					}
				}
			} else {
				img, _ := c.Adapter.Library.Get(n.Properties.Id + "canvas")
				width, height := n.GetStyle("width"), n.GetStyle("height")
				if n.GetStyle("width") == "" {
					width = strconv.Itoa(img.Bounds().Dx()) + "px"
				}
				if n.GetStyle("height") == "" {
					height = strconv.Itoa(img.Bounds().Dy()) + "px"
				}
				n.SetStyle("width", width)
				n.SetStyle("height", height)
				n.Canvas = &canvas.Canvas{
					RGBA: img,
				}
			}

			return n
		},
	}
}
