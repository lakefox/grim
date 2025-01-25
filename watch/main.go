package watch

import "grim/element"

// there already is a render change name or move grimRender to here and call this from grim.Path
func Check(n element.Node, change string, values ...string) {
	// if something like (n, "style", "width", "20px")
	// then update the n.Style and check parent to see if it fits and how everything
	// needs to update and do that
	// do we need two documents? the only thing I can see is that .Style would have all styles
	// but say border: 2px solid red; and change style border-left-width 10px
	// then just rebuild the border selector to include it bc we already break all selectors
}
