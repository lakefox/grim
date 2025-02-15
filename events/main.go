package events

import (
	adapter "grim/adapters"
	"grim/cstyle"
	"grim/element"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

type EventData struct {
	Position []int
	Click    bool
	Context  bool
	ScrollX  int
	ScrollY  int
	Key      int
	KeyState bool
}

type Monitor struct {
	Adapter  *adapter.Adapter
	EventMap map[string]element.Event
	CSS      *cstyle.CSS
	Focus    Focus
	Drag     Drag
}

type Drag struct {
	Position []int
	Type     string
}

type Focus struct {
	Selected            int
	LastClickWasFocused bool
	Nodes               []string
	SoftFocused         string
}

// !ISSUE: Events fire multiple times find a way to prevent sending repeat events at 120 fps also during drag click gets fired for every element moused over

// !TODO: Should return modified elements, take nothing as a input bc m.CSS.Document
// + need to find what changed, remove GetStyles might have to do after adding
// + style precompute
// + reomve m.EventMap after done if possible
// + for k,v := range m./EventMap
// + prob storing computed styles should be first bc then you can tell if the event matters
func (m *Monitor) RunEvents(n *element.Node) bool {
	var scrolled bool
	for _, v := range n.Children {
		r := m.RunEvents(v)
		if r {
			scrolled = r
		}
	}

	evt := m.EventMap[n.Properties.Id]
	evt.Target = n

	if scrolled {
		evt.ScrollX = 0
		evt.ScrollY = 0
		m.EventMap[n.Properties.Id] = evt
	}
	eventListeners := []string{}

	if evt.MouseDown {
		if n.OnMouseDown != nil {
			n.OnMouseDown(evt)
		}
		eventListeners = append(eventListeners, "mousedown")
	}

	if evt.MouseUp {
		if n.OnMouseUp != nil {
			n.OnMouseUp(evt)
		}
		eventListeners = append(eventListeners, "mouseup")
	}

	if evt.Click {
		if n.OnClick != nil {
			n.OnClick(evt)
		}
		eventListeners = append(eventListeners, "click")
	}

	if evt.ContextMenu {
		if n.OnContextMenu != nil {
			n.OnContextMenu(evt)
		}
		eventListeners = append(eventListeners, "contextmenu")
	}

	if evt.MouseEnter {
		if n.OnMouseEnter != nil {
			n.OnMouseEnter(evt)
		}
		eventListeners = append(eventListeners, "mouseenter")
	}

	if evt.MouseOver {
		if n.OnMouseOver != nil {
			n.OnMouseOver(evt)
		}
		eventListeners = append(eventListeners, "mouseover")
	}

	if evt.MouseLeave {
		if n.OnMouseLeave != nil {
			n.OnMouseLeave(evt)
		}
		eventListeners = append(eventListeners, "mouseleave")
	}

	if evt.MouseMove {
		if n.OnMouseLeave != nil {
			n.OnMouseLeave(evt)
		}
		eventListeners = append(eventListeners, "mouseleave")
	}

	if evt.Hover {
		n.Hovered = true
	} else {
		n.Hovered = false
	}

	if len(m.Focus.Nodes) > 0 && m.Focus.Selected > -1 {
		if m.Focus.Nodes[m.Focus.Selected] == n.Properties.Id {
			n.Focus()
		} else {
			n.Blur()
		}
	} else {
		n.Blur()
	}

	var styledEl map[string]string

	if evt.ScrollX != 0 || evt.ScrollY != 0 {
		styledEl = n.Styles()
	}

	if evt.ScrollX != 0 {
		if hasAutoOrScroll(styledEl) {
			s := m.CSS.State
			self := s[n.Properties.Id]
			containerWidth := self.Width
			n.ScrollLeft += evt.ScrollX

			if (int((float32(int(n.ScrollLeft))/((containerWidth/float32(n.ScrollWidth))*containerWidth))*containerWidth) + int(containerWidth)) >= n.ScrollWidth {
				n.ScrollLeft = (((n.ScrollWidth) - int(containerWidth)) * int(containerWidth)) / n.ScrollWidth
			}

			if n.ScrollLeft <= 0 {
				n.ScrollLeft = 0
			}

			if n.OnScroll != nil {
				n.OnScroll(evt)
			}

			evt.ScrollX = 0
			m.EventMap[n.Properties.Id] = evt
			scrolled = true
		}
	}

	if evt.ScrollY != 0 {
		if hasAutoOrScroll(styledEl) {
			s := m.CSS.State
			self := s[n.Properties.Id]
			containerHeight := self.Height
			n.ScrollTop -= evt.ScrollY

			// This is the scroll scaling equation if it is less than the scroll height then let it add the next scroll amount
			if (int((float32(int(n.ScrollTop))/((containerHeight/float32(n.ScrollHeight))*containerHeight))*containerHeight) + int(containerHeight)) >= n.ScrollHeight {
				n.ScrollTop = (((n.ScrollHeight) - int(containerHeight)) * int(containerHeight)) / n.ScrollHeight
			}

			if n.ScrollTop <= 0 {
				n.ScrollTop = 0
			}

			if n.OnScroll != nil {
				n.OnScroll(evt)
			}

			evt.ScrollY = 0
			m.EventMap[n.Properties.Id] = evt
			scrolled = true
		}
	}

	for _, v := range eventListeners {
		if len(n.Properties.EventListeners[v]) > 0 {
			for _, handler := range n.Properties.EventListeners[v] {
				handler(evt)
			}
		}
	}
	return scrolled
}

type fn struct {
	Id       string
	TabIndex int
}

func (m *Monitor) GetEvents(data *EventData) {
	headElements := []string{
		"head",
		"title",    // Defines the title of the document
		"base",     // Specifies the base URL for all relative URLs in the document
		"link",     // Links to external resources like stylesheets
		"meta",     // Provides metadata about the document (e.g., character set, viewport)
		"style",    // Embeds internal CSS styles
		"script",   // Embeds or references JavaScript code
		"noscript", // Provides alternate content for users without JavaScript
		"template", // Used to define a client-side template
	}

	s := m.CSS.State

	m.Focus.LastClickWasFocused = false
	// update focesable nodes
	nodes := []fn{}
	for k, self := range s {
		if self.TabIndex > -1 {
			nodes = append(nodes, fn{Id: k, TabIndex: self.TabIndex})
		}
	}

	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].TabIndex < nodes[j].TabIndex // Ascending order by TabIndex
	})

	m.Focus.Nodes = []string{}
	for _, v := range nodes {
		good := true

		for _, tag := range headElements {
			if len(v.Id) >= len(tag) {
				if v.Id[0:len(tag)] == tag {
					good = false
				}
			}
		}

		if good {
			m.Focus.Nodes = append(m.Focus.Nodes, v.Id)
		}
	}
	if m.Drag.Position == nil {
		m.Drag = Drag{Position: []int{-1, -1}}
	}
	drag := false
	if m.Drag.Position != nil {
		if m.Drag.Position[0] > -1 && m.Drag.Position[1] > -1 {
			// !ISSUE: works Y only also it should only fire on only draggable
			drag = true
		}
	}

	if data.Position == nil {
		return
	}

	var softFocus string

	for k, self := range s {
		var isMouseOver, isFocused bool

		if m.Focus.Selected > -1 {
			isFocused = m.Focus.Nodes[m.Focus.Selected] == k
		} else if m.Focus.SoftFocused != "" {
			isFocused = m.Focus.SoftFocused == k
		} else {
			isFocused = false
		}

		evt, ok := m.EventMap[k]

		if !ok {
			evt = element.Event{}
		}

		insideX := (self.X < float32(data.Position[0]) && self.X+self.Width > float32(data.Position[0]))
		insideY := (self.Y < float32(data.Position[1]) && self.Y+self.Height > float32(data.Position[1]))
		inside := (insideX && insideY)

		arrowScrollX := 0
		arrowScrollY := 0

		if m.Focus.SoftFocused == k || inside {
			if data.Key == 265 {
				// up
				arrowScrollY += 20
			} else if data.Key == 264 {
				// Down
				arrowScrollY -= 20
			}
			if data.Key == 262 {
				// up
				arrowScrollX += 20
			} else if data.Key == 263 {
				// Down
				arrowScrollX -= 20
			}
		}

		if isFocused {

			// This allows for arrow scrolling when off the element and typing

			// Get the keycode of the pressed key
			if data.Key != 0 {
				if self.ContentEditable {
					// Sync the innertext and value but idk
					// !ISSUE: This may not work
					ProcessKeyEvent(self, int(data.Key))

					evt.Value = self.Value
				}
			}

			if data.Key == 258 && data.KeyState && !m.Focus.LastClickWasFocused {
				// Tab
				mfsLen := len(m.Focus.Nodes)
				if mfsLen > 0 {
					store := m.Focus.Selected
					m.Focus.Selected += 1
					if m.Focus.Selected >= mfsLen {
						m.Focus.Selected = 0
					}
					if store != m.Focus.Selected {
						m.Focus.LastClickWasFocused = true
					}
				}

			}
		}

		if inside || isFocused {
			// Mouse is over element
			isMouseOver = true

			evt.Hover = true

			if data.Click && !evt.MouseDown {
				evt.MouseDown = true
				evt.MouseUp = false
				if m.Drag.Position[0] == -1 && m.Drag.Position[1] == -1 {
					if strings.Contains(k, "grim-thumb") {
						t := ""
						if strings.Contains(k, "grim-thumbx") {
							t = "x"
						} else if strings.Contains(k, "grim-thumby") {
							t = "y"
						}
						m.Drag = Drag{Position: data.Position,
							Type: t}
					}
				}
			}

			if !data.Click && !evt.MouseUp {
				evt.MouseUp = true
				evt.MouseDown = false
				evt.Click = false
				m.Drag = Drag{Position: []int{-1, -1}}
			}

			if data.Click && !evt.Click && !drag {
				evt.Click = true

				if !inside && !(self.TabIndex > -1) {
					if m.Focus.SoftFocused == k {
						m.Focus.SoftFocused = ""
						softFocus = ""
					}
				}

				if self.TabIndex > -1 {
					if m.Focus.Selected > -1 {
						if len(m.Focus.Nodes) > 0 && m.Focus.Selected > -1 {
							if m.Focus.Nodes[m.Focus.Selected] != k {
								for i, v := range m.Focus.Nodes {
									if v == k {
										m.Focus.Selected = i
										m.Focus.LastClickWasFocused = true
										break
									}
								}
							} else {
								m.Focus.LastClickWasFocused = true
							}
						} else {
							m.Focus.LastClickWasFocused = true
						}

					} else {
						selectedIndex := -1
						for i, v := range m.Focus.Nodes {
							if v == k {
								selectedIndex = i
							}
						}
						if selectedIndex == -1 {
							nodes = append(nodes, fn{Id: k, TabIndex: self.TabIndex})
							sort.Slice(nodes, func(i, j int) bool {
								return nodes[i].TabIndex < nodes[j].TabIndex // Ascending order by TabIndex
							})
							m.Focus.Nodes = []string{}
							for _, v := range nodes {
								m.Focus.Nodes = append(m.Focus.Nodes, v.Id)
							}

							for i, v := range m.Focus.Nodes {
								if v == k {
									selectedIndex = i
								}
							}
						}

						m.Focus.Selected = selectedIndex
						m.Focus.LastClickWasFocused = true
					}

				} else if m.Focus.Selected > -1 {
					if len(m.Focus.Nodes) > 0 && m.Focus.Selected > -1 {
						if m.Focus.Nodes[m.Focus.Selected] != k && !m.Focus.LastClickWasFocused {
							m.Focus.Selected = -1
						}
					}
				}

				if inside && m.Focus.Selected == -1 {
					if softFocus == "" {
						softFocus = k
					} else {
						if s[softFocus].Z < s[k].Z {
							softFocus = k
						} else if s[softFocus].Z == s[k].Z {
							if extractNumber(k) < extractNumber(softFocus) {
								softFocus = k
							}
						}
					}
				}

			}
			// Regardless set soft focus to trigger events to the selected element: when non is set default body???

			if data.Context {
				evt.ContextMenu = true
			}

			if (data.ScrollY != 0 && (inside)) || (data.ScrollX != 0 && (inside)) || arrowScrollX != 0 || arrowScrollY != 0 || drag {
				if drag {
					if m.Drag.Type == "y" {
						data.ScrollY = (evt.Y - data.Position[1])
					} else if m.Drag.Type == "x" {
						data.ScrollX = -(evt.X - data.Position[0])
					}
				}
				evt.ScrollX = data.ScrollX + arrowScrollX
				evt.ScrollY = data.ScrollY + arrowScrollY
				arrowScrollX = 0
				arrowScrollY = 0
				if strings.Contains(k, "grim-thumb") || drag {
					data.ScrollX = 0
					data.ScrollY = 0
				}
			}

			if !evt.MouseEnter && inside {
				evt.MouseEnter = true
				evt.MouseOver = true
				evt.MouseLeave = false

				// Let the adapter know the cursor has changed
				m.Adapter.DispatchEvent(element.Event{
					Name: "cursor",
					Data: self.Cursor,
				})
			}

			if inside {
				evt.MouseMove = true
				evt.X = data.Position[0]
				evt.Y = data.Position[1]
			} else {
				evt.MouseMove = true
			}

		} else {
			isMouseOver = false
			evt.Hover = false
		}

		if !isMouseOver && !evt.MouseLeave {
			evt.MouseEnter = false
			evt.MouseOver = false
			evt.MouseLeave = true
			// n.Properties.Hover = false
		}

		if evt.X != int(data.Position[0]) && evt.Y != int(data.Position[1]) {
			evt.X = int(data.Position[0])
			evt.Y = int(data.Position[1])
		}

		m.EventMap[k] = evt
	}

	if softFocus != "" {
		m.Focus.SoftFocused = softFocus
	}
}

// ProcessKeyEvent processes key events for text entry.
func ProcessKeyEvent(self element.State, key int) {
	// Handle key events for text entry
	switch key {
	case 8:
		// Backspace: remove the last character
		if len(self.Value) > 0 {
			self.Value = self.Value[:len(self.Value)-1]
			// n.InnerText = n.InnerText[:len(n.InnerText)-1]
		}

	case 65:
		// !TODO: ctrl a
		// // Select All: set the entire text as selected
		// if key == 17 || key == 345 {
		// 	n.Properties.Selected = []float32{0, float32(len(n.Value))}
		// } else {
		// 	// Otherwise, append 'A' to the text
		// 	n.Value += "A"
		// }

	case 67:
		// Copy: copy the selected text (in this case, print it)
		// if key == 17 || key == 345 {
		// } else {
		// 	// Otherwise, append 'C' to the text
		// 	n.Value += "C"
		// }

	case 86:
		// Paste: paste the copied text (in this case, set it to "Pasted")
		// if key == 17 || key == 345 {
		// 	n.Value = "Pasted"
		// } else {
		// 	// Otherwise, append 'V' to the text
		// 	n.Value += "V"
		// }

	default:
		// Record other key presses
		self.Value += string(rune(key))
	}
}

// !ISSUE: Optiomise this to just take the element
func hasAutoOrScroll(styledEl map[string]string) bool {
	overflowKeys := []string{"overflow", "overflow-x", "overflow-y"}
	for _, key := range overflowKeys {
		if value, exists := styledEl[key]; exists {
			values := strings.Fields(value) // Splits the value by spaces
			for _, v := range values {
				if v == "auto" || v == "scroll" {
					return true
				}
			}
		}
	}
	return false
}
func extractNumber(input string) int {
	var numStr string
	for _, char := range input {
		if unicode.IsDigit(char) {
			numStr += string(char)
		}
	}
	if numStr == "" {
		return 0
	}
	n, _ := strconv.Atoi(numStr)
	return n
}
