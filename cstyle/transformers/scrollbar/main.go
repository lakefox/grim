package scrollbar

import (
	"grim/cstyle"
	"grim/element"
	"strconv"
	"strings"
)

func Init() cstyle.Transformer {
	return cstyle.Transformer{
		Selector: func(n *element.Node) bool {
			if n.Style["overflow"] != "" || n.Style["overflow-x"] != "" || n.Style["overflow-y"] != "" {
				return true
			} else {
				return false
			}
		},
		Handler: func(n *element.Node, c *cstyle.CSS) *element.Node {
			overflowProps := strings.Split(n.Style["overflow"], " ")
			if n.Style["overflow-y"] == "" {
				val := overflowProps[0]
				if len(overflowProps) >= 2 {
					val = overflowProps[1]
				}
				n.Style["overflow-y"] = val
			}
			if n.Style["overflow-x"] == "" {
				n.Style["overflow-x"] = overflowProps[0]
			}

			if n.Style["position"] == "" {
				n.Style["position"] = "relative"
			}

			trackWidth := "14px"
			thumbWidth := "8px"
			thumbMargin := "3px"
			if n.Style["scrollbar-width"] == "thin" {
				trackWidth = "10px"
				thumbWidth = "6px"
				thumbMargin = "2px"
			}
			if n.Style["scrollbar-width"] == "none" {
				return n
			}

			splitStr := strings.Split(n.Style["scrollbar-color"], " ")

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

			if (n.Style["overflow-x"] == "scroll" || n.Style["overflow-x"] == "auto") && ((n.ScrollWidth > int(c.Width) && n.TagName == "html") || n.ScrollWidth > 0) {
				scrollbar := n.CreateElement("grim-track")

				scrollbar.Style["position"] = "absolute"
				scrollbar.Style["bottom"] = "0px"
				scrollbar.Style["left"] = "0"
				scrollbar.Style["width"] = "100%"
				scrollbar.Style["height"] = trackWidth
				scrollbar.Style["background"] = backgroundColor
				scrollbar.SetAttribute("direction", "x")

				thumb := n.CreateElement("grim-thumbx")
				thumb.Style["position"] = "absolute"
				thumb.Style["left"] = strconv.Itoa(n.ScrollLeft) + "px"
				thumb.Style["top"] = thumbMargin
				thumb.Style["height"] = thumbWidth
				thumb.Style["width"] = "20px"
				thumb.Style["background"] = thumbColor
				thumb.Style["cursor"] = "pointer"
				thumb.Style["border-radius"] = "10px"

				for k, v := range n.PseudoElements["::-webkit-scrollbar"] {
					scrollbar.Style[k] = v
					thumb.Style[k] = v
				}

				for k, v := range n.PseudoElements["::-webkit-scrollbar-track"] {
					scrollbar.Style[k] = v
				}

				for k, v := range n.PseudoElements["::-webkit-scrollbar-thumb"] {
					thumb.Style[k] = v
				}

				scrollbar.AppendChild(&thumb)
				xScrollbar = true
				n.AppendChild(&scrollbar)
			}

			// Y scrollbar

			if (n.Style["overflow-y"] == "scroll" || n.Style["overflow-y"] == "auto") && ((n.ScrollHeight > int(c.Height) && n.TagName == "html") || n.ScrollHeight > 0) {
				scrollbar := n.CreateElement("grim-track")

				scrollbar.Style["position"] = "absolute"
				scrollbar.Style["top"] = "0"
				scrollbar.Style["right"] = "0"
				scrollbar.Style["width"] = trackWidth
				scrollbar.Style["height"] = "100%"
				scrollbar.Style["background"] = backgroundColor
				scrollbar.SetAttribute("direction", "y")

				if xScrollbar {
					scrollbar.Style["height"] = "calc(100% - " + trackWidth + ")"
				}

				thumb := n.CreateElement("grim-thumby")

				thumb.Style["position"] = "absolute"
				thumb.Style["top"] = strconv.Itoa(n.ScrollTop) + "px"
				thumb.Style["left"] = "0"
				thumb.Style["width"] = thumbWidth
				thumb.Style["height"] = "20px"
				thumb.Style["background"] = thumbColor
				thumb.Style["cursor"] = "pointer"
				thumb.Style["margin-left"] = thumbMargin
				thumb.Style["border-radius"] = "10px"

				for k, v := range n.PseudoElements["::-webkit-scrollbar"] {
					scrollbar.Style[k] = v
					thumb.Style[k] = v
				}

				for k, v := range n.PseudoElements["::-webkit-scrollbar-track"] {
					scrollbar.Style[k] = v
				}

				for k, v := range n.PseudoElements["::-webkit-scrollbar-thumb"] {
					thumb.Style[k] = v
				}

				scrollbar.AppendChild(&thumb)

				n.Style["width"] = "calc(" + n.Style["width"] + "-" + trackWidth + ")"
				pr := n.Style["padding-right"]
				if pr == "" {
					if n.Style["padding"] != "" {
						pr = n.Style["padding"]
					}
				}

				if pr != "" {
					n.Style["padding-right"] = "calc(" + pr + "+" + trackWidth + ")"
				} else {
					n.Style["padding-right"] = trackWidth
				}
				n.AppendChild(&scrollbar)
			}

			return n
		},
	}
}
