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
	window := grim.New(raylib.Init())
	window.Path("./src/index.html")

	document := window.Document

	fmt.Println(document.QuerySelector(":nth-child(1)"))

	grim.Open(&window, 850, 400)
}
