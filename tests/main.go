package main

import (
	"fmt"
	"grim"
	"grim/adapters/raylib"
)

// go tool pprof --pdf ./main ./cpu.pprof > file.pdf && open file.pdf
// go tool pprof --pdf ./main ./mem.pprof > file.pdf && open file.pdf

func main() {
	// defer profile.Start(profile.ProfilePath(".")).Stop() // CPU
	// defer profile.Start(profile.MemProfile, profile.ProfilePath(".")).Stop() // Memory
	// defaults read ~/Library/Preferences/.GlobalPreferences.plist
	// !ISSUE: Flex2 doesn't work anymore
	window := grim.New(raylib.Init(), 850, 400)
	window.Path("./src/index.html")
	document := window.Document()

	// !ISSUE: Code is unreachable
	// qsa := document.QuerySelectorAll(`:where(h1, h2, h3)`)
	qsa := document.QuerySelector(`ul`)

	for _, v := range qsa.Children {
		fmt.Println("qsa: ", v.OuterHTML())
	}
	// body:has(h1.class + h1)> h1[attr="test"]#id.class:has(input:is(input[type="text"]) + div),a {
}
