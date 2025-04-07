package main

import (
	"fmt"
	"grim"
	"grim/adapters/raylib"
	// "net/http"
	// _ "net/http/pprof"
)

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
	window.Path("./src/index.html")
	document := window.Document()

	// qsa := document.QuerySelectorAll(`:where(h1, h2, h3)`)
	qsa := document.QuerySelector(`.box`)
	fmt.Println(qsa.ClassList)

	window.Open()

}
