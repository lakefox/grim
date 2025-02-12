package ol

import (
	"fmt"
	"grim/cstyle"
	"grim/element"
	"grim/font"
	"grim/utils"
	"strconv"

	imgFont "golang.org/x/image/font"
)

func Init() cstyle.Transformer {
	return cstyle.Transformer{
		Selector: func(n *element.Node, c *cstyle.CSS) bool {
			return n.TagName == "ol"
		},
		Handler: func(n *element.Node, c *cstyle.CSS) *element.Node {
			tN := n.CreateElement(n.TagName)
			var maxOS int
			var widths []int
			for i, v := range n.Children {
				li := n.CreateElement("li")
				dot := li.CreateElement("div")
				content := li.CreateElement("div")

				for k, v := range v.Styles() {
					li.Style(k, v)
					dot.Style(k, v)
					content.Style(k, v)
				}

				li.Style("display", "flex")
				li.Style("align-items", "center")

				dot.Style("margin-right", "6px")
				dot.Style("display", "block")

				italic := false

				if n.Style("font-style") == "italic" {
					italic = true
				}

				if c.Fonts == nil {
					c.Fonts = map[string]imgFont.Face{}
				}

				fs := utils.ConvertToPixels(n.Style("font-size"), 16, c.Width)
				em := fs

				fid := n.Style("font-family") + fmt.Sprint(em, n.Style("font-weight"), italic)
				if c.Fonts[fid] == nil {
					f, err := font.LoadFont(n.Style("font-family"), int(em), n.Style("font-weight"), italic, &c.Adapter.FileSystem)

					if err != nil {
						panic(err)
					}
					c.Fonts[fid] = f
				}
				fnt := c.Fonts[fid]
				w, _ := font.MeasureText(&font.MetaData{Font: &fnt}, strconv.Itoa(i+1)+".")
				widths = append(widths, w)
				if w > maxOS {
					maxOS = w
				}

				dot.InnerText = strconv.Itoa(i+1) + "."

				content.InnerText = v.InnerText
				content.Style("display", "block")

				li.AppendChild(&dot)
				li.AppendChild(&content)
				li.Parent = n

				tN.AppendChild(&li)
			}

			for i := range tN.Children {
				tN.Children[i].Children[0].Style("margin-left", strconv.Itoa((maxOS-widths[i]))+"px")
			}
			n.Children = tN.Children
			return n
		},
	}
}
