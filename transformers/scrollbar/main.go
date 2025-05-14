package scrollbar

import (
	"grim"
	"strconv"
	"strings"
)

func Init() grim.Transformer {
	return grim.Transformer{
		Selector: func(n *grim.Node, c *grim.CSS) bool {
			style := n.Styles()
			if style["overflow"] != "" || style["overflow-x"] != "" || style["overflow-y"] != "" {
				return true
			} else {
				return false
			}
		},
		Handler: func(n *grim.Node, c *grim.CSS) *grim.Node {
			style := n.Styles()
			top, left := n.GetScroll()
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
				n.AppendChild(&scrollbar)

				scrollbar.SetStyle("position", "absolute")
				scrollbar.SetStyle("bottom", "0px")
				scrollbar.SetStyle("left", "0")
				scrollbar.SetStyle("width", "100%")
				scrollbar.SetStyle("height", trackWidth)
				scrollbar.SetStyle("background-color", backgroundColor)
				scrollbar.SetStyle("z-index", "99999")
				scrollbar.SetAttribute("direction", "x")

				thumb := n.CreateElement("grim-thumbx")
				scrollbar.AppendChild(&thumb)

				thumb.SetStyle("position", "absolute")
				thumb.SetStyle("left", strconv.Itoa(left) + "px")
				thumb.SetStyle("top", thumbMargin)
				thumb.SetStyle("height", thumbWidth)
				thumb.SetStyle("width", "20px")
				thumb.SetStyle("background-color", thumbColor)
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
			}

			// Y scrollbar

			if n.GetStyle("overflow-y") == "scroll" || n.GetStyle("overflow-y") == "auto" {
				scrollbar := n.CreateElement("grim-track")
				n.AppendChild(&scrollbar)

				scrollbar.SetStyle("position", "absolute")
				scrollbar.SetStyle("top", "0")
				scrollbar.SetStyle("right", "0")
				scrollbar.SetStyle("width", trackWidth)
				scrollbar.SetStyle("height", "100%")
				scrollbar.SetStyle("background-color", backgroundColor)
				scrollbar.SetStyle("z-index", "99999")
				scrollbar.SetAttribute("direction", "y")

				thumb := n.CreateElement("grim-thumby")
				scrollbar.AppendChild(&thumb)

				thumb.SetStyle("position", "absolute")
				thumb.SetStyle("top", strconv.Itoa(top) + "px")
				// !ISSUE: parse the string then calculate the offset for thin and normal
				thumb.SetStyle("right", "3px")
				thumb.SetStyle("width", thumbWidth)
				thumb.SetStyle("height", "20px")
				thumb.SetStyle("background-color", thumbColor)
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

				// !NOTE: This prevents recursion
				if !strings.Contains(style["width"], "calc") {
					n.SetStyle("width", "calc(" + style["width"] + "-" + trackWidth + ")")
				}

				pr := n.GetStyle("padding-right")
				// !ISSUE: calc() should not be used
				if pr == "" && n.GetStyle("padding") != "" {
					n.SetStyle("padding-right", "calc(" + n.InitalStyles["padding"] + " + " + trackWidth + ")")
				} else if pr != "" {
					n.SetStyle("padding-right", "calc(" + n.InitalStyles["padding-right"] + " + " + trackWidth + ")")
				} else {
					n.SetStyle("padding-right", trackWidth)
				}
			}

			return n
		},
	}
}
