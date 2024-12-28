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

			width := "14px"
			if n.Style["scrollbar-width"] == "thin" {
				width = "10px"
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
				backgroundColor = "#fafafa"
				thumbColor = "#c7c7c7"

			}

			// !TODO: Make X scroll bar

			// Y scrollbar

			if (n.Style["overflow-y"] == "scroll" || n.Style["overflow-y"] == "auto") && ((n.ScrollHeight > int(c.Height) && n.TagName == "html") || n.ScrollHeight > 0) {
				scrollbar := n.CreateElement("grim-track")

				scrollbar.Style["position"] = "absolute"
				scrollbar.Style["top"] = "0"
				scrollbar.Style["right"] = "0"
				scrollbar.Style["width"] = width
				scrollbar.Style["height"] = "100%"
				scrollbar.Style["background"] = backgroundColor

				thumb := n.CreateElement("grim-thumb")

				thumb.Style["position"] = "absolute"
				thumb.Style["top"] = strconv.Itoa(n.ScrollTop) + "px"
				thumb.Style["left"] = "0"
				thumb.Style["width"] = width
				thumb.Style["height"] = "20px"
				thumb.Style["background"] = thumbColor
				thumb.Style["cursor"] = "pointer"

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

				n.Style["width"] = "calc(" + n.Style["width"] + "-" + width + ")"
				pr := n.Style["padding-right"]
				if pr == "" {
					if n.Style["padding"] != "" {
						pr = n.Style["padding"]
					}
				}

				if pr != "" {
					n.Style["padding-right"] = "calc(" + pr + "+" + width + ")"
				} else {
					n.Style["padding-right"] = width
				}

				n.AppendChild(&scrollbar)
			}

			return n
		},
	}
}
