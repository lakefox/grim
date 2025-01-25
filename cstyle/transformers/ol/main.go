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
		Selector: func(n *element.Node) bool {
			return n.TagName == "ol"
		},
		Handler: func(n *element.Node, c *cstyle.CSS) *element.Node {
			tN := n.CreateElement(n.TagName)
			var maxOS int
			var widths []int
			for i, v := range n.Children {
				li := n.CreateElement("li")

				li.CStyle = v.CStyle
				li.CStyle["display"] = "flex"
				li.CStyle["align-items"] = "center"
				li.CStyle = c.QuickStyles(&li)

				dot := li.CreateElement("div")
				dot.CStyle = li.CStyle
				dot.CStyle["margin-right"] = "6px"

				dot.CStyle = c.QuickStyles(&dot)
				dot.CStyle["display"] = "block"

				italic := false

				if n.CStyle["font-style"] == "italic" {
					italic = true
				}

				if c.Fonts == nil {
					c.Fonts = map[string]imgFont.Face{}
				}

				fs := utils.ConvertToPixels(n.CStyle["font-size"], 16, c.Width)
				em := fs

				fid := n.CStyle["font-family"] + fmt.Sprint(em, n.CStyle["font-weight"], italic)
				if c.Fonts[fid] == nil {
					f, _ := font.LoadFont(n.CStyle["font-family"], int(em), n.CStyle["font-weight"], italic, &c.Adapter.FileSystem)
					c.Fonts[fid] = f
				}
				fnt := c.Fonts[fid]
				w, _ := font.MeasureText(&font.MetaData{Font: &fnt}, strconv.Itoa(i+1)+".")
				widths = append(widths, w)
				if w > maxOS {
					maxOS = w
				}

				dot.InnerText = strconv.Itoa(i+1) + "."

				content := li.CreateElement("div")
				content.InnerText = v.InnerText
				content.CStyle = li.CStyle
				content.CStyle = c.QuickStyles(&content)
				content.CStyle["display"] = "block"

				li.AppendChild(&dot)
				li.AppendChild(&content)
				li.Parent = n

				tN.AppendChild(&li)
			}

			for i := range tN.Children {
				tN.Children[i].Children[0].CStyle["margin-left"] = strconv.Itoa((maxOS - widths[i])) + "px"
			}
			n.Children = tN.Children
			return n
		},
	}
}
