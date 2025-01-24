package main

import (
	"fmt"
	"grim"
	"grim/adapters/raylib"
	"grim/element"
)

// go tool pprof --pdf ./main ./cpu.pprof > file.pdf && open file.pdf
// go tool pprof --pdf ./main ./mem.pprof > file.pdf && open file.pdf

func main() {
	// defer profile.Start(profile.ProfilePath(".")).Stop() // CPU
	// defer profile.Start(profile.MemProfile, profile.ProfilePath(".")).Stop() // Memory
	// defaults read ~/Library/Preferences/.GlobalPreferences.plist
	// !ISSUE: Flex2 doesn't work anymore
	window := grim.New(raylib.Init())
	window.Path("./src/index.html")
	document := window.Document()
	fmt.Println("start")
	fmt.Println(element.ExtractBaseElements(`:where(h1, h2, h3)`))
	fmt.Println("end")
	// qsa := document.QuerySelectorAll(`:where(h1, h2, h3)`)
	qsa := document.QuerySelectorAll(`:where(h1, h2, h3)`)

	for _, v := range *qsa {
		fmt.Println("qsa: ", v.Properties.Id)
	}
	// body:has(h1.class + h1)> h1[attr="test"]#id.class:has(input:is(input[type="text"]) + div),a {
	grim.Open(&window, 850, 400)
}
