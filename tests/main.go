package main

import (
	"grim"
	"grim/adapters/raylib"
	// "github.com/pkg/profile"
)

// go tool pprof --pdf ./main ./cpu.pprof > file.pdf && open file.pdf
// go tool pprof --pdf ./main ./mem.pprof > file.pdf && open file.pdf

func main() {
	// defer profile.Start(profile.ProfilePath(".")).Stop() // CPU
	// defer profile.Start(profile.MemProfile, profile.ProfilePath(".")).Stop() // Memory
	// defaults read ~/Library/Preferences/.GlobalPreferences.plist
	// !ISSUE: Flex2 doesn't work anymore
	window := grim.New(raylib.Init())

	window.Path("./src/superselector.html")

	grim.Open(&window, 850, 400)
}
