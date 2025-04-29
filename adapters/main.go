package adapter

import (
	"grim/element"
	"image"
)

type Adapter struct {
	Init       func(width int, height int)
	Render     func(state []element.State)
	Load       func(key string, texture image.Image)
	Unload     func(key string)
	events     map[string][]func(element.Event)
	FileSystem FileSystem
	// id -> type -> key
	Textures map[string]map[string]string
}

func Init() {}

// Only render complex
func RenderType() {}

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

// !ISSUE: Make a init function
func (a *Adapter) LoadTexture(id, t, key string, texture image.Image) {
	a.Load(key, texture)
	if a.Textures == nil {
		a.Textures = map[string]map[string]string{}
	}
	if a.Textures[id] == nil {
		a.Textures[id] = map[string]string{}
	}
	a.Textures[id][t] = key
}

func (a *Adapter) UnloadTexture(id, t string) {
	a.Unload(a.Textures[id][t])
	delete(a.Textures[id], t)
	if len(a.Textures[id]) == 0 {
		delete(a.Textures[id], t)
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
