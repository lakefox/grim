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
			if n.CStyle["overflow"] != "" || n.CStyle["overflow-x"] != "" || n.CStyle["overflow-y"] != "" {
				return true
			} else {
				return false
			}
		},
		Handler: func(n *element.Node, c *cstyle.CSS) *element.Node {
			overflowProps := strings.Split(n.CStyle["overflow"], " ")
			if n.CStyle["overflow-y"] == "" {
				val := overflowProps[0]
				if len(overflowProps) >= 2 {
					val = overflowProps[1]
				}
				n.CStyle["overflow-y"] = val
			}
			if n.CStyle["overflow-x"] == "" {
				n.CStyle["overflow-x"] = overflowProps[0]
			}

			if n.CStyle["position"] == "" {
				n.CStyle["position"] = "relative"
			}

			trackWidth := "14px"
			thumbWidth := "8px"
			thumbMargin := "3px"
			if n.CStyle["scrollbar-width"] == "thin" {
				trackWidth = "10px"
				thumbWidth = "6px"
				thumbMargin = "2px"
			}
			if n.CStyle["scrollbar-width"] == "none" {
				return n
			}

			splitStr := strings.Split(n.CStyle["scrollbar-color"], " ")

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

			if n.CStyle["overflow-x"] == "scroll" || n.CStyle["overflow-x"] == "auto" {
				scrollbar := n.CreateElement("grim-track")

				scrollbar.CStyle["position"] = "absolute"
				scrollbar.CStyle["bottom"] = "0px"
				scrollbar.CStyle["left"] = "0"
				scrollbar.CStyle["width"] = "100%"
				scrollbar.CStyle["height"] = trackWidth
				scrollbar.CStyle["background"] = backgroundColor
				scrollbar.CStyle["z-index"] = "99999"
				scrollbar.SetAttribute("direction", "x")

				thumb := n.CreateElement("grim-thumbx")
				thumb.CStyle["position"] = "absolute"
				thumb.CStyle["left"] = strconv.Itoa(n.ScrollLeft) + "px"
				thumb.CStyle["top"] = thumbMargin
				thumb.CStyle["height"] = thumbWidth
				thumb.CStyle["width"] = "20px"
				thumb.CStyle["background"] = thumbColor
				thumb.CStyle["cursor"] = "pointer"
				thumb.CStyle["border-radius"] = "10px"
				thumb.CStyle["z-index"] = "99999"

				for k, v := range n.PseudoElements["::-webkit-scrollbar"] {
					scrollbar.CStyle[k] = v
					thumb.CStyle[k] = v
				}

				for k, v := range n.PseudoElements["::-webkit-scrollbar-track"] {
					scrollbar.CStyle[k] = v
				}

				for k, v := range n.PseudoElements["::-webkit-scrollbar-thumb"] {
					thumb.CStyle[k] = v
				}

				scrollbar.Properties.Id = element.GenerateUniqueId(n, scrollbar.TagName)
				scrollbar.AppendChild(&thumb)
				xScrollbar = true
				n.AppendChild(&scrollbar)
			}

			// Y scrollbar

			if n.CStyle["overflow-y"] == "scroll" || n.CStyle["overflow-y"] == "auto" {
				scrollbar := n.CreateElement("grim-track")

				scrollbar.CStyle["position"] = "absolute"
				scrollbar.CStyle["top"] = "0"
				scrollbar.CStyle["right"] = "0"
				scrollbar.CStyle["width"] = trackWidth
				scrollbar.CStyle["height"] = "100%"
				scrollbar.CStyle["background"] = backgroundColor
				scrollbar.CStyle["z-index"] = "99999"
				scrollbar.SetAttribute("direction", "y")

				if xScrollbar {
					scrollbar.CStyle["height"] = "calc(100% - " + trackWidth + ")"
				}

				thumb := n.CreateElement("grim-thumby")

				thumb.CStyle["position"] = "absolute"
				thumb.CStyle["top"] = strconv.Itoa(n.ScrollTop) + "px"
				thumb.CStyle["left"] = "0"
				thumb.CStyle["width"] = thumbWidth
				thumb.CStyle["height"] = "20px"
				thumb.CStyle["background"] = thumbColor
				thumb.CStyle["cursor"] = "pointer"
				thumb.CStyle["margin-left"] = thumbMargin
				thumb.CStyle["border-radius"] = "10px"
				thumb.CStyle["z-index"] = "99999"

				for k, v := range n.PseudoElements["::-webkit-scrollbar"] {
					scrollbar.CStyle[k] = v
					thumb.CStyle[k] = v
				}

				for k, v := range n.PseudoElements["::-webkit-scrollbar-track"] {
					scrollbar.CStyle[k] = v
				}

				for k, v := range n.PseudoElements["::-webkit-scrollbar-thumb"] {
					thumb.CStyle[k] = v
				}
				scrollbar.Properties.Id = element.GenerateUniqueId(n, scrollbar.TagName)
				scrollbar.AppendChild(&thumb)

				n.CStyle["width"] = "calc(" + n.CStyle["width"] + "-" + trackWidth + ")"
				pr := n.CStyle["padding-right"]
				if pr == "" {
					if n.CStyle["padding"] != "" {
						pr = n.CStyle["padding"]
					}
				}

				if pr != "" {
					n.CStyle["padding-right"] = "calc(" + pr + "+" + trackWidth + ")"
				} else {
					n.CStyle["padding-right"] = trackWidth
				}
				n.AppendChild(&scrollbar)
			}

			return n
		},
	}
}
