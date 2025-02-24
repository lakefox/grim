package scrollbar

import (
	"grim/cstyle"
	"grim/element"
	"grim/utils"
	"strconv"
	"strings"
)

func Init() cstyle.Transformer {
	return cstyle.Transformer{
		Selector: func(n *element.Node, c *cstyle.CSS) bool {
			style := n.Styles()
			if style["overflow"] != "" || style["overflow-x"] != "" || style["overflow-y"] != "" {
				return true
			} else {
				return false
			}
		},
		Handler: func(n *element.Node, c *cstyle.CSS) *element.Node {
			style := n.Styles()

			overflowProps := strings.Split(style["overflow"], " ")
			if n.Style("overflow-y") == "" {
				val := overflowProps[0]
				if len(overflowProps) >= 2 {
					val = overflowProps[1]
				}
				n.Style("overflow-y", val)
			}
			if style["overflow-x"] == "" {
				n.Style("overflow-x", overflowProps[0])
			}

			if style["position"] == "" {
				n.Style("position", "relative")
			}

			trackWidth := "14px"
			thumbWidth := "8px"
			thumbMargin := "3px"
			if n.Style("scrollbar-width") == "thin" {
				trackWidth = "10px"
				thumbWidth = "6px"
				thumbMargin = "2px"
			}
			if n.Style("scrollbar-width") == "none" {
				return n
			}

			ps := n.StyleSheets.PsuedoStyles[n.Properties.Id]

			splitStr := strings.Split(n.Style("scrollbar-color"), " ")

			// Initialize the variables
			var backgroundColor, thumbColor string

			// Check the length of the split result and assign the values accordingly

			if len(splitStr) >= 2 {
				backgroundColor = splitStr[1]
				thumbColor = splitStr[0]
			} else {
				backgroundColor = "#2b2b2b"
				thumbColor = "#6b6b6b"

			}

			// X scrollbar
			xScrollbar := false

			if n.Style("overflow-x") == "scroll" || n.Style("overflow-x") == "auto" {
				scrollbar := n.CreateElement("grim-track")

				scrollbar.Style("position", "absolute")
				scrollbar.Style("bottom", "0px")
				scrollbar.Style("left", "0")
				scrollbar.Style("width", "100%")
				scrollbar.Style("height", trackWidth)
				scrollbar.Style("background", backgroundColor)
				scrollbar.Style("z-index", "99999")
				scrollbar.SetAttribute("direction", "x")

				thumb := n.CreateElement("grim-thumbx")
				thumb.Style("position", "absolute")
				thumb.Style("left", strconv.Itoa(n.ScrollLeft)+"px")
				thumb.Style("top", thumbMargin)
				thumb.Style("height", thumbWidth)
				thumb.Style("width", "20px")
				thumb.Style("background", thumbColor)
				thumb.Style("cursor", "pointer")
				thumb.Style("border-radius", "10px")
				thumb.Style("z-index", "99999")

				for k, v := range ps["::-webkit-scrollbar"] {
					scrollbar.Style(k, v)
					thumb.Style(k, v)
				}

				for k, v := range ps["::-webkit-scrollbar-track"] {
					scrollbar.Style(k, v)
				}

				for k, v := range ps["::-webkit-scrollbar-thumb"] {
					thumb.Style(k, v)
				}

				scrollbar.Properties.Id = element.GenerateUniqueId(n, scrollbar.TagName)
				scrollbar.AppendChild(&thumb)
				xScrollbar = true
				n.AppendChild(&scrollbar)
			}

			// Y scrollbar

			if n.Style("overflow-y") == "scroll" || n.Style("overflow-y") == "auto" {
				scrollbar := n.CreateElement("grim-track")

				scrollbar.Style("position", "absolute")
				scrollbar.Style("top", "0")
				scrollbar.Style("right", "0")
				scrollbar.Style("width", trackWidth)
				scrollbar.Style("height", "100%")
				scrollbar.Style("background", backgroundColor)
				scrollbar.Style("z-index", "99999")
				scrollbar.SetAttribute("direction", "y")

				if xScrollbar {
					scrollbar.Style("height", "calc(100% - "+trackWidth+")")
				}

				thumb := n.CreateElement("grim-thumby")

				thumb.Style("position", "absolute")
				thumb.Style("top", strconv.Itoa(n.ScrollTop)+"px")
				// !ISSUE: parse the string then calculate the offset for thin and normal
				thumb.Style("right", "3px")
				thumb.Style("width", thumbWidth)
				thumb.Style("height", "20px")
				thumb.Style("background", thumbColor)
				thumb.Style("cursor", "pointer")
				thumb.Style("margin-left", thumbMargin)
				thumb.Style("border-radius", "10px")
				thumb.Style("z-index", "99999")

				for k, v := range ps["::-webkit-scrollbar"] {
					scrollbar.Style(k, v)
					thumb.Style(k, v)
				}

				for k, v := range ps["::-webkit-scrollbar-track"] {
					scrollbar.Style(k, v)
				}

				for k, v := range ps["::-webkit-scrollbar-thumb"] {
					thumb.Style(k, v)
				}
				scrollbar.Properties.Id = element.GenerateUniqueId(n, scrollbar.TagName)
				scrollbar.AppendChild(&thumb)

				// !DEVMAN,NOTE: This prevents recursion
				if !strings.Contains(style["width"], "calc") {
					n.Style("width", "calc("+style["width"]+"-"+trackWidth+")")
				}

				pr := n.Style("padding-right")

				if pr == "" && n.Style("padding") != "" {
					n.Style("padding-right", "calc("+n.StyleSheets.Styles["padding"]+" + "+trackWidth+")")
				} else if n.Style("padding") != "" {
					_,r,_,_ := utils.ConvertMarginToIndividualProperties(n.Style("padding"))
					n.Style("padding-right", "calc("+r+" + "+trackWidth+")")
				} else {
					n.Style("padding-right", trackWidth)
				}

				n.AppendChild(&scrollbar)
			}

			return n
		},
	}
}
