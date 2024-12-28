package adapter

import (
	"grim/element"
	"grim/library"
	"os"
	"path/filepath"
	"sort"
)

type Adapter struct {
	Init       func(width int, height int)
	Render     func(state []element.State)
	Load       func(state []element.State)
	events     map[string][]func(element.Event)
	Library    *library.Shelf
	FileSystem FileSystem
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

type FileSystem struct {
	Paths     []string
	ReadFile  func(path string) ([]byte, error)
	WriteFile func(path string, data []byte)
}

func (fs *FileSystem) AddFile(path string) {
	fs.Paths = append(fs.Paths, path)
}

func (fs *FileSystem) AddDir(path string) error {
	// Walk through the directory and collect all file paths
	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() { // Only add files, not directories
			fs.Paths = append(fs.Paths, filePath)
		}
		return nil
	})
	return err
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
func (fs *FileSystem) FindFile(query string) string {
	var fileDistances []fileWithDistance

	// Calculate the Levenshtein distance for each file path
	for _, path := range fs.Paths {
		distance := Levenshtein(query, filepath.Base(path))
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
