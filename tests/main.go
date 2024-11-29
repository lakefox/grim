package main

import (
	"gui"
	"gui/adapters/raylib"
	// "github.com/pkg/profile"
)

// go tool pprof --pdf ./main ./cpu.pprof > file.pdf && open file.pdf
// go tool pprof --pdf ./main ./mem.pprof > file.pdf && open file.pdf

func main() {
	// defer profile.Start(profile.ProfilePath(".")).Stop() // CPU
	// defer profile.Start(profile.MemProfile, profile.ProfilePath(".")).Stop() // Memory
	// defaults read ~/Library/Preferences/.GlobalPreferences.plist
	// !ISSUE: Flex2 doesn't work anymore
	window := gui.Open("./src/index.html", raylib.Init())

	// document := window.Document

	// body := document.QuerySelector("body")

	// body.AddEventListener("click", func(e element.Event) {
	// 	fmt.Println("click", e.Target.OuterHTML)
	// })

	// tgt(body)

	// document.QuerySelector("body").AddEventListener("scroll", func(e element.Event) {
	// 	fmt.Println(e.Target.ScrollY, e.Target.TagName)
	// })

	// canvas := document.CreateElement("canvas")
	// canvas.Style["background"] = "#00f"
	// ctx := canvas.GetContext(300, 300)

	// ctx.BeginPath()
	// // ctx.MoveTo(0, 0)
	// // ctx.LineTo(100, 100)
	// ctx.LineWidth = 10
	// ctx.RoundedRect(10, 10, 100, 100, []int{50, 40})
	// ctx.FillStyle = color.RGBA{255, 0, 0, 255}
	// ctx.StrokeStyle = color.RGBA{255, 0, 0, 255}
	// ctx.Stroke()
	// ctx.ClosePath()
	// body.AppendChild(&canvas)

	gui.View(&window, 850, 400)
}

// func tgt(e *element.Node) {
// 	// events need to be transfered to broke out elements
// 	e.AddEventListener("click", func(e element.Event) {
// 		// fmt.Println(document.QuerySelector("body").InnerHTML)
// 		fmt.Println(e.Target.TagName)
// 		fmt.Println(e.Target.InnerHTML)
// 		// e.Target.Style["background"] = "red"
// 	})

// 	for i := range e.Children {

// 		tgt(e.Children[i])
// 	}
// }
