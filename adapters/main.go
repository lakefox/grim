package adapter

import (
	"grim/element"
	"grim/library"
	"os"
	"path/filepath"
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
	Sources   []string
	ReadFile  func(path string) ([]byte, error)
	WriteFile func(path string, data []byte)
}

func (fs *FileSystem) AddFile(path string) {
	fs.Paths = append(fs.Paths, path)
}

func (fs *FileSystem) AddDir(path string) error {
	// Walk through the directory and collect all file paths
	fs.Sources = append(fs.Sources, path)
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

