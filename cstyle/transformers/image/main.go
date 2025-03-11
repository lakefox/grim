package img

import (
	// "bytes"
	"grim/cstyle"
	"grim/element"
	// "image"
	// _ "image/gif"  // Enable GIF support
	// _ "image/jpeg" // Enable JPEG support
	// _ "image/png"  // Enable PNG support
	// "path/filepath"
	// "strconv"
)

func Init() cstyle.Transformer {
	return cstyle.Transformer{
		Selector: func(n *element.Node, c *cstyle.CSS) bool {
			return n.TagName == "img"
			// !ISSUE: img tags or background-url || background: url()
		},
		Handler: func(n *element.Node, c *cstyle.CSS) *element.Node {
			n.TagName = "canvas"
			// res, exists := c.Adapter.Textures[n.Properties.Id]["background"]
			// if !exists {
			// 	i, err := c.Adapter.FileSystem.ReadFile(filepath.Join(c.Path, n.Src))
			// 	if err == nil {
			// 		img, _, err := image.Decode(bytes.NewReader(i))
			// 		if err == nil {
			// 			width, height := n.ComputedStyle["width"], n.ComputedStyle["height"]
			// 			if n.ComputedStyle["width"] == "" {
			// 				width = strconv.Itoa(img.Bounds().Dx()) + "px"
			// 			}
			// 			if n.ComputedStyle["height"] == "" {
			// 				height = strconv.Itoa(img.Bounds().Dy()) + "px"
			// 			}
			// 			ctx := n.GetContext(img.Bounds().Dx(), img.Bounds().Dy())
			// 			n.ComputedStyle["width"] = width
			// 			n.ComputedStyle["height"] = height
			// 			ctx.DrawImage(img, 0, 0)
			// 		}
			// 	}
			// } else {
			// 	img, _ := c.Adapter.Library.Get(n.Properties.Id + "canvas")
			// 	width, height := n.ComputedStyle["width"], n.ComputedStyle["height"]
			// 	if width == "" {
			// 		width = strconv.Itoa(img.Bounds().Dx()) + "px"
			// 	}
			// 	if height == "" {
			// 		height = strconv.Itoa(img.Bounds().Dy()) + "px"
			// 	}
			// 	// !ISSUE: Doesn't resize and uses a lot of memory (nah its the top)
			// 	// + here resize the image then store it, then just add the key to the state.Textures
			// 	ctx := n.GetContext(img.Bounds().Dx(), img.Bounds().Dy())
			// 	n.ComputedStyle["width"] = width
			// 	n.ComputedStyle["height"] = height
			// 	ctx.DrawImage(img, 0, 0)
			// }

			return n
		},
	}
}
