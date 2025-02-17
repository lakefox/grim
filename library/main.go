package library

import (
	"image"
)

// Borrow checker for images

type Shelf struct {
	Textures       map[string]image.RGBA
	References     map[string]bool
	UnloadCallback func(string)
}

func (s *Shelf) Set(key string, img image.RGBA) string {
	if s.Textures == nil {
		s.Textures = map[string]image.RGBA{}
	}
	if s.References == nil {
		s.References = map[string]bool{}
	}

	s.Textures[key] = img
	s.References[key] = true
	return key
}

func (s *Shelf) Get(key string) (*image.RGBA, bool) {
	a, exists := s.Textures[key]
	return &a, exists
}

// Check marks the reference as true if the texture exists.
func (s *Shelf) Check(key string) bool {
	if _, exists := s.Textures[key]; exists {
		s.References[key] = true
		return true
	}
	return false
}

func (s *Shelf) Bounds(key string) []int {
	if _, exists := s.Textures[key]; exists {
		img := s.Textures[key]
		b := img.Bounds()
		return []int{b.Dx(), b.Dy()}
	} else {
		return []int{0, 0}
	}
}

func (s *Shelf) Delete(key string) {
	delete(s.Textures, key)
}

func (s *Shelf) Clean() {
	for k, v := range s.References {
		if !v {
			if s.UnloadCallback != nil {
				s.UnloadCallback(k)
			}
			delete(s.References, k)
			delete(s.Textures, k)
		} else {
			// Only reset the reference if it was true
			s.References[k] = false
		}
	}
}
