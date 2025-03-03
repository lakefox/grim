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
			if n.GetStyle("overflow-y") == "" {
				val := overflowProps[0]
				if len(overflowProps) >= 2 {
					val = overflowProps[1]
				}
				n.SetStyle("overflow-y", val)
			}
			if style["overflow-x"] == "" {
				n.SetStyle("overflow-x", overflowProps[0])
			}

			if style["position"] == "" {
				n.SetStyle("position", "relative")
			}

			trackWidth := "14px"
			thumbWidth := "8px"
			thumbMargin := "3px"
			if n.GetStyle("scrollbar-width") == "thin" {
				trackWidth = "10px"
				thumbWidth = "6px"
				thumbMargin = "2px"
			}
			if n.GetStyle("scrollbar-width") == "none" {
				return n
			}

			ps := n.StyleSheets.PsuedoStyles[n.Properties.Id]

			splitStr := strings.Split(n.GetStyle("scrollbar-color"), " ")

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
			if n.GetStyle("overflow-x") == "scroll" || n.GetStyle("overflow-x") == "auto" {
				scrollbar := n.CreateElement("grim-track")

				scrollbar.SetStyle("position", "absolute")
				scrollbar.SetStyle("bottom", "0px")
				scrollbar.SetStyle("left", "0")
				scrollbar.SetStyle("width", "100%")
				scrollbar.SetStyle("height", trackWidth)
				scrollbar.SetStyle("background", backgroundColor)
				scrollbar.SetStyle("z-index", "99999")
				scrollbar.SetAttribute("direction", "x")

				thumb := n.CreateElement("grim-thumbx")
				thumb.SetStyle("position", "absolute")
				thumb.SetStyle("left", strconv.Itoa(n.ScrollLeft)+"px")
				thumb.SetStyle("top", thumbMargin)
				thumb.SetStyle("height", thumbWidth)
				thumb.SetStyle("width", "20px")
				thumb.SetStyle("background", thumbColor)
				thumb.SetStyle("cursor", "pointer")
				thumb.SetStyle("border-radius", "10px")
				thumb.SetStyle("z-index", "99999")

				for k, v := range ps["::-webkit-scrollbar"] {
					scrollbar.SetStyle(k, v)
					thumb.SetStyle(k, v)
				}

				for k, v := range ps["::-webkit-scrollbar-track"] {
					scrollbar.SetStyle(k, v)
				}

				for k, v := range ps["::-webkit-scrollbar-thumb"] {
					thumb.SetStyle(k, v)
				}

				scrollbar.Properties.Id = element.GenerateUniqueId(n, scrollbar.TagName)
				scrollbar.AppendChild(&thumb)
				n.AppendChild(&scrollbar)
			}

			// Y scrollbar

			if n.GetStyle("overflow-y") == "scroll" || n.GetStyle("overflow-y") == "auto" {
				scrollbar := n.CreateElement("grim-track")

				scrollbar.SetStyle("position", "absolute")
				scrollbar.SetStyle("top", "0")
				scrollbar.SetStyle("right", "0")
				scrollbar.SetStyle("width", trackWidth)
				scrollbar.SetStyle("height", "100%")
				scrollbar.SetStyle("background", backgroundColor)
				scrollbar.SetStyle("z-index", "99999")
				scrollbar.SetAttribute("direction", "y")

				thumb := n.CreateElement("grim-thumby")

				thumb.SetStyle("position", "absolute")
				thumb.SetStyle("top", strconv.Itoa(n.ScrollTop)+"px")
				// !ISSUE: parse the string then calculate the offset for thin and normal
				thumb.SetStyle("right", "3px")
				thumb.SetStyle("width", thumbWidth)
				thumb.SetStyle("height", "20px")
				thumb.SetStyle("background", thumbColor)
				thumb.SetStyle("cursor", "pointer")
				thumb.SetStyle("margin-left", thumbMargin)
				thumb.SetStyle("border-radius", "10px")
				thumb.SetStyle("z-index", "99999")

				for k, v := range ps["::-webkit-scrollbar"] {
					scrollbar.SetStyle(k, v)
					thumb.SetStyle(k, v)
				}

				for k, v := range ps["::-webkit-scrollbar-track"] {
					scrollbar.SetStyle(k, v)
				}

				for k, v := range ps["::-webkit-scrollbar-thumb"] {
					thumb.SetStyle(k, v)
				}
				scrollbar.Properties.Id = element.GenerateUniqueId(n, scrollbar.TagName)
				scrollbar.AppendChild(&thumb)

				// !DEVMAN,NOTE: This prevents recursion
				if !strings.Contains(style["width"], "calc") {
					n.SetStyle("width", "calc("+style["width"]+"-"+trackWidth+")")
				}

				pr := n.GetStyle("padding-right")

				if pr == "" && n.GetStyle("padding") != "" {
					n.SetStyle("padding-right", "calc("+n.StyleSheets.Styles["padding"]+" + "+trackWidth+")")
				} else if n.GetStyle("padding") != "" {
					_, r, _, _ := utils.ConvertMarginToIndividualProperties(n.GetStyle("padding"))
					n.SetStyle("padding-right", "calc("+r+" + "+trackWidth+")")
				} else {
					n.SetStyle("padding-right", trackWidth)
				}
				n.AppendChild(&scrollbar)
			}

			return n
		},
	}
}
