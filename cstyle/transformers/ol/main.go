package ol

import (
	"fmt"
	"grim/cstyle"
	"grim/element"
	"grim/font"
	"grim/utils"
	"strconv"

	"github.com/golang/freetype/truetype"
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
			// !ISSUE: Update this to match ul
			for i, v := range n.Children {
				li := n.CreateElement("li")
				dot := li.CreateElement("div")
				content := li.CreateElement("div")

				for k, v := range v.ComputedStyle {
					li.ComputedStyle[k] = v
					dot.ComputedStyle[k] = v
					content.ComputedStyle[k] = v
				}

				li.ComputedStyle["display"] = "flex"
				li.ComputedStyle["align-items"] = "center"

				dot.ComputedStyle["margin-right"] = "6px"
				dot.ComputedStyle["display"] = "block"

				italic := false

				if n.ComputedStyle["font-style"] == "italic" {
					italic = true
				}

				if c.Fonts == nil {
					c.Fonts = map[string]*truetype.Font{}
				}

				fs := utils.ConvertToPixels(n.ComputedStyle["font-size"], 16, c.Width)
				em := fs

				fid := n.ComputedStyle["font-family"] + fmt.Sprint(em, n.ComputedStyle["font-weight"], italic)
				fnt, ok := c.Fonts[fid]

				if !ok {
					f, err := font.LoadFont(n.ComputedStyle["font-family"], int(em), n.ComputedStyle["font-weight"], italic, &c.Adapter.FileSystem)

					if err != nil {
						panic(err)
					}
					c.Fonts[fid] = f
				}

				w := font.MeasureText(&font.MetaData{Font: fnt}, strconv.Itoa(i+1)+".")
				widths = append(widths, w)
				if w > maxOS {
					maxOS = w
				}

				dot.InnerText = strconv.Itoa(i+1) + "."

				content.InnerText = v.InnerText
				content.ComputedStyle["display"] = "block"

				li.AppendChild(&dot)
				li.AppendChild(&content)
				li.Parent = n

				tN.AppendChild(&li)
			}

			for i := range tN.Children {
				tN.Children[i].Children[0].ComputedStyle["margin-left"] = strconv.Itoa((maxOS - widths[i])) + "px"
			}
			n.Children = tN.Children
			return n
		},
	}
}
