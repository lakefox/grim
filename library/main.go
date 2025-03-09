package library

import (
	"fmt"
	"image"
)

// Borrow checker for images

type Shelf struct {
	Textures       map[string]image.Image
	Bounds         map[string]Bounds
	References     map[string]bool
	UnloadCallback func(string)
}

type Bounds struct {
	Width  int
	Height int
}

func (s *Shelf) Set(key string, img image.Image) string {
	if s.Textures == nil {
		s.Textures = map[string]image.Image{}
	}
	if s.References == nil {
		s.References = map[string]bool{}
	}
	if s.Bounds == nil {
		s.Bounds = map[string]Bounds{}
	}
	b := img.Bounds()
	s.Bounds[key] = Bounds{Width: b.Dx(), Height: b.Dy()}
	s.Textures[key] = img
	s.References[key] = true
	fmt.Println("set",key)
	return key
}

func (s *Shelf) Get(key string) (image.Image, bool) {
	a, exists := s.Textures[key]
	s.Textures[key] = nil
	return a, exists
}

// Check marks the reference as true if the texture exists.
func (s *Shelf) Check(key string) bool {
	if _, exists := s.Textures[key]; exists {
		s.References[key] = true
		return true
	}
	return false
}

func (s *Shelf) GetBounds(key string) (Bounds, bool) {
	b, exists := s.Bounds[key]
	// s.References[key] = true
	return b, exists
}

func (s *Shelf) Delete(key string) {
	delete(s.Textures, key)
	delete(s.Bounds, key)
	delete(s.References, key)
	fmt.Println("delete", key)
}

func (s *Shelf) Clean() {
	for k, v := range s.References {
		if !v {
			if s.UnloadCallback != nil {
				s.UnloadCallback(k)
			}
			delete(s.References, k)
			delete(s.Textures, k)
			delete(s.Bounds, k)
			fmt.Println("clean",k)
		} else {
			// Only reset the reference if it was true
			s.References[k] = false
		}
	}
}
