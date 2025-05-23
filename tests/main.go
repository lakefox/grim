package main

import (
	"fmt"
	"grim"
	"grim/adapters/raylib"
	"grim/plugins/crop"
	"grim/plugins/flex"
	"grim/plugins/inline"
	"grim/plugins/textAlign"
	"grim/scripts/a"
	"grim/transformers/banda"

	marginblock "grim/transformers/margin-block"
	"grim/transformers/ol"
	"grim/transformers/scrollbar"
	"grim/transformers/text"
	"grim/transformers/ul"
	// "net/http"
	// _ "net/http/pprof"
)

// go tool pprof -http=localhost:5678 http://localhost:6060/debug/pprof/hea
// func logMemoryUsage() {
// 	var memStats runtime.MemStats
//
// 	for {
// 		runtime.ReadMemStats(&memStats)
// 		fmt.Printf("Time: %s\n", time.Now().Format("15:04:05"))
// 		fmt.Printf("Alloc: %v KB\n", memStats.Alloc/1024)               // Memory allocated and still in use
// 		fmt.Printf("TotalAlloc: %v KB\n", memStats.TotalAlloc/1024)     // Total memory allocated (even if freed)
// 		fmt.Printf("Sys: %v KB\n", memStats.Sys/1024)                   // Total memory obtained from the OS
// 		fmt.Printf("HeapAlloc: %v KB\n", memStats.HeapAlloc/1024)       // Heap memory currently allocated
// 		fmt.Printf("HeapSys: %v KB\n", memStats.HeapSys/1024)           // Heap memory reserved from the OS
// 		fmt.Printf("HeapIdle: %v KB\n", memStats.HeapIdle/1024)         // Heap memory not currently in use
// 		fmt.Printf("HeapInuse: %v KB\n", memStats.HeapInuse/1024)       // Heap memory currently in use
// 		fmt.Printf("HeapReleased: %v KB\n", memStats.HeapReleased/1024) // Heap memory returned to the OS
// 		fmt.Printf("HeapObjects: %v\n", memStats.HeapObjects)           // Number of allocated objects
// 		fmt.Println("---------------------------------------------------")
// 		time.Sleep(1 * time.Second) // Log every second
// 	}
// }

func main() {
	// go func() {
	// 	http.ListenAndServe("localhost:6060", nil)
	// }()
	// go logMemoryUsage()
	// !ISSUE: Flex2 doesn't work anymore
	window := grim.New(raylib.Init(), 850, 400)

	window.Plugins(inline.Init(), textAlign.Init(), flex.Init(), crop.Init())
	window.Transformers(text.Init(), banda.Init(), scrollbar.Init(), marginblock.Init(), ul.Init(), ol.Init())
	window.Scripts(a.Init())

	window.Path("./src/index.html")
	document := window.Document()

	// qsa := document.QuerySelectorAll(`:where(h1, h2, h3)`)
	qsa := document.QuerySelector(".box")
	fmt.Println("Classlist: ", qsa.ClassList)

	window.Open()

}
