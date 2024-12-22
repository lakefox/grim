package adapter

import (
	"gui/element"
	"gui/library"
	"sort"
)

// !TODO: For things like fonts the file system interaction works fine, but if you want to run grim on a pi pico it doesn't have the file system int like
// + computers have
// + Option 1: make a file system adapter where authors can control how data is read
// + Option 2: at build time fetch all files needed and bundle them

type Adapter struct {
	Init      func(width int, height int)
	Render    func(state []element.State)
	Load      func(state []element.State)
	events    map[string][]func(element.Event)
	Library   *library.Shelf
	Options   Options
	FilePaths []string
}

type Options struct {
	RenderText     bool
	RenderElements bool
	RenderBorders  bool
}

func (a *Adapter) AddEventListener(name string, callback func(element.Event)) {
	if a.events == nil {
		a.events = map[string][]func(element.Event){}
	}
	a.events[name] = append(a.events[name], callback)
}

func (a *Adapter) DispatchEvent(event element.Event) {
	if a.events != nil {
		evts := a.events[event.Name]
		for _, v := range evts {
			v(event)
		}
	}
}

func (a *Adapter) AddFilePath(path string) {
	a.FilePaths = append(a.FilePaths, path)
}

// Levenshtein Distance Function (no external packages)
func Levenshtein(a, b string) int {
	lenA, lenB := len(a), len(b)
	// Create a matrix to hold distances
	matrix := make([][]int, lenA+1)
	for i := range matrix {
		matrix[i] = make([]int, lenB+1)
	}

	// Initialize the first row and first column
	for i := 0; i <= lenA; i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= lenB; j++ {
		matrix[0][j] = j
	}

	// Calculate the Levenshtein distance
	for i := 1; i <= lenA; i++ {
		for j := 1; j <= lenB; j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			matrix[i][j] = min(matrix[i-1][j]+1, matrix[i][j-1]+1, matrix[i-1][j-1]+cost)
		}
	}

	return matrix[lenA][lenB]
}

// Helper function to return the minimum of three values
func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// Struct to hold file path and its Levenshtein distance
type fileWithDistance struct {
	path     string
	distance int
}

// Fuzzy search function that returns the closest matches, sorted by Levenshtein distance
func (a *Adapter) FindFile(query string) string {
	var fileDistances []fileWithDistance

	// Calculate the Levenshtein distance for each file path
	for _, path := range a.FilePaths {
		distance := Levenshtein(query, path)
		fileDistances = append(fileDistances, fileWithDistance{path, distance})
	}

	// Sort the fileDistances slice by distance (ascending order)
	sort.Slice(fileDistances, func(i, j int) bool {
		return fileDistances[i].distance < fileDistances[j].distance
	})

	// Create a slice to hold only the sorted file paths
	var sortedPaths []string
	for _, file := range fileDistances {
		sortedPaths = append(sortedPaths, file.path)
	}

	return sortedPaths[0]
}
