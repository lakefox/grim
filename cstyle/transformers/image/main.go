package img

import (
	"bytes"
	"grim/cstyle"
	"grim/element"
	"image"
	_ "image/gif"  // Enable GIF support
	_ "image/jpeg" // Enable JPEG support
	_ "image/png"  // Enable PNG support
	"path/filepath"
)

func Init() cstyle.Transformer {
	return cstyle.Transformer{
		Selector: func(n *element.Node, c *cstyle.CSS) bool {
			return n.TagName == "img"
			// !ISSUE: img tags or background-url || background: url()
		},
		Handler: func(n *element.Node, c *cstyle.CSS) *element.Node {
			key, exists := c.Adapter.Textures[n.Properties.Id]["image"]
			file := filepath.Join(c.Path, n.Src)
			if !exists || key != file {
				if exists {
					c.Adapter.UnloadTexture(n.Properties.Id, "image")
				}
				i, err := c.Adapter.FileSystem.ReadFile(file)
				if err == nil {
					img, _, err := image.Decode(bytes.NewReader(i))
					if err == nil {
						c.Adapter.LoadTexture(n.Properties.Id, "image", file, img)
					}
				}
			}

			return n
		},
	}
}
