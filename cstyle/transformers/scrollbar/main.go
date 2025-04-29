package scrollbar

import (
	"grim/cstyle"
	"grim/element"
	"strconv"
	"strings"
)

func Init() cstyle.Transformer {
	return cstyle.Transformer{
		Selector: func(n *element.Node, c *cstyle.CSS) bool {
			style := n.ComputedStyle
			if style["overflow"] != "" || style["overflow-x"] != "" || style["overflow-y"] != "" {
				return true
			} else {
				return false
			}
		},
		Handler: func(n *element.Node, c *cstyle.CSS) *element.Node {
			style := n.ComputedStyle

			overflowProps := strings.Split(style["overflow"], " ")
			if n.ComputedStyle["overflow-y"] == "" {
				val := overflowProps[0]
				if len(overflowProps) >= 2 {
					val = overflowProps[1]
				}
				n.ComputedStyle["overflow-y"] = val
			}
			if style["overflow-x"] == "" {
				n.ComputedStyle["overflow-x"] = overflowProps[0]
			}

			if style["position"] == "" {
				n.ComputedStyle["position"] = "relative"
			}

			trackWidth := "14px"
			thumbWidth := "8px"
			thumbMargin := "3px"
			if n.ComputedStyle["scrollbar-width"] == "thin" {
				trackWidth = "10px"
				thumbWidth = "6px"
				thumbMargin = "2px"
			}
			if n.ComputedStyle["scrollbar-width"] == "none" {
				return n
			}

			ps := n.StyleSheets.PsuedoStyles[n.Properties.Id]

			splitStr := strings.Split(n.ComputedStyle["scrollbar-color"], " ")

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
			if n.ComputedStyle["overflow-x"] == "scroll" || n.ComputedStyle["overflow-x"] == "auto" {
				scrollbar := n.CreateElement("grim-track")

				scrollbar.SetStyle("position", "absolute")
				scrollbar.SetStyle("bottom", "0px")
				scrollbar.SetStyle("left", "0")
				scrollbar.SetStyle("width", "100%")
				scrollbar.SetStyle("height", trackWidth)
				scrollbar.SetStyle("background-color", backgroundColor)
				scrollbar.SetStyle("z-index", "99999")
				scrollbar.SetAttribute("direction", "x")

				thumb := n.CreateElement("grim-thumbx")
				thumb.ComputedStyle["position"] = "absolute"
				thumb.ComputedStyle["left"] = strconv.Itoa(n.ScrollLeft) + "px"
				thumb.ComputedStyle["top"] = thumbMargin
				thumb.ComputedStyle["height"] = thumbWidth
				thumb.ComputedStyle["width"] = "20px"
				thumb.ComputedStyle["background-color"] = thumbColor
				thumb.ComputedStyle["cursor"] = "pointer"
				thumb.ComputedStyle["border-radius"] = "10px"
				thumb.ComputedStyle["z-index"] = "99999"

				for k, v := range ps["::-webkit-scrollbar"] {
					scrollbar.ComputedStyle[k] = v
					thumb.ComputedStyle[k] = v
				}

				for k, v := range ps["::-webkit-scrollbar-track"] {
					scrollbar.ComputedStyle[k] = v
				}

				for k, v := range ps["::-webkit-scrollbar-thumb"] {
					thumb.ComputedStyle[k] = v
				}

				scrollbar.Properties.Id = element.GenerateUniqueId(n, scrollbar.TagName())
				scrollbar.AppendChild(&thumb)
				n.AppendChild(&scrollbar)
			}

			// Y scrollbar

			if n.ComputedStyle["overflow-y"] == "scroll" || n.ComputedStyle["overflow-y"] == "auto" {
				scrollbar := n.CreateElement("grim-track")

				scrollbar.SetStyle("position", "absolute")
				scrollbar.SetStyle("top", "0")
				scrollbar.SetStyle("right", "0")
				scrollbar.SetStyle("width", trackWidth)
				scrollbar.SetStyle("height", "100%")
				scrollbar.SetStyle("background-color", backgroundColor)
				scrollbar.SetStyle("z-index", "99999")
				scrollbar.SetAttribute("direction", "y")

				thumb := n.CreateElement("grim-thumby")

				thumb.ComputedStyle["position"] = "absolute"
				thumb.ComputedStyle["top"] = strconv.Itoa(n.ScrollTop) + "px"
				// !ISSUE: parse the string then calculate the offset for thin and normal
				thumb.ComputedStyle["right"] = "3px"
				thumb.ComputedStyle["width"] = thumbWidth
				thumb.ComputedStyle["height"] = "20px"
				thumb.ComputedStyle["background-color"] = thumbColor
				thumb.ComputedStyle["cursor"] = "pointer"
				thumb.ComputedStyle["margin-left"] = thumbMargin
				thumb.ComputedStyle["border-radius"] = "10px"
				thumb.ComputedStyle["z-index"] = "99999"

				for k, v := range ps["::-webkit-scrollbar"] {
					scrollbar.ComputedStyle[k] = v
					thumb.ComputedStyle[k] = v
				}

				for k, v := range ps["::-webkit-scrollbar-track"] {
					scrollbar.ComputedStyle[k] = v
				}

				for k, v := range ps["::-webkit-scrollbar-thumb"] {
					thumb.ComputedStyle[k] = v
				}

				// !NOTE: This need to be here because at this point scrollbar hasn't been added to the parent
				// + ,so it doesn't have a prid to build off of
				scrollbar.Properties.Id = element.GenerateUniqueId(n, scrollbar.TagName())
				scrollbar.AppendChild(&thumb)

				// !NOTE: This prevents recursion
				if !strings.Contains(style["width"], "calc") {
					n.ComputedStyle["width"] = "calc(" + style["width"] + "-" + trackWidth + ")"
				}

				pr := n.ComputedStyle["padding-right"]
				// !ISSUE: remove appendchild
				if pr == "" && n.ComputedStyle["padding"] != "" {
					n.ComputedStyle["padding-right"] = "calc(" + n.InitalStyles["padding"] + " + " + trackWidth + ")"
				} else if pr != "" {
					n.ComputedStyle["padding-right"] = "calc(" + n.InitalStyles["padding-right"] + " + " + trackWidth + ")"
				} else {
					n.ComputedStyle["padding-right"] = trackWidth
				}
				n.AppendChild(&scrollbar)
			}

			return n
		},
	}
}

