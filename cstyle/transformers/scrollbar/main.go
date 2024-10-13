package scrollbar

import (
	"fmt"
	"gui/cstyle"
	"gui/element"
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

			fmt.Println(n.ScrollHeight, n.Style["height"])

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

			// Y scrollbar

			if n.Style["overflow-y"] == "scroll" || n.Style["overflow-y"] == "auto" {
				scrollbar := n.CreateElement("grim-scrollbar")

				scrollbar.Style["position"] = "absolute"
				scrollbar.Style["top"] = "0"
				scrollbar.Style["right"] = "0"
				scrollbar.Style["width"] = width
				scrollbar.Style["height"] = "100%"
				scrollbar.Style["z-index"] = "9"
				scrollbar.Style["background"] = backgroundColor

				thumb := n.CreateElement("grim-thumb")

				thumb.Style["position"] = "absolute"
				thumb.Style["top"] = strconv.Itoa(n.ScrollTop) + "px"
				thumb.Style["left"] = "0"
				thumb.Style["width"] = width
				thumb.Style["height"] = "20px"
				thumb.Style["background"] = thumbColor
				thumb.Style["cursor"] = "pointer"
				thumb.Style["z-index"] = "10"
				scrollbar.AppendChild(&thumb)

				// !ISSUE: need to remove the scroll bar if there is no need for it
				// + overflow-y: auto would be the only case... but need to see if the contents overflow
				// + currently can't think of a way to do this but n.ScrollHeight does exist
				// + just don't know the height of the container if say height: 100%. that is calculated later
				// + but it needs to be removed here
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
