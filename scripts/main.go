package scripts

import "grim/element"

type Scripts struct {
	scripts []Script
}

type Script struct {
	Call func(*element.Node)
}

func (s *Scripts) Run(n *element.Node) {
	for _, v := range s.scripts {
		v.Call(n)
	}
}

func (s *Scripts) Add(c Script) {
	s.scripts = append(s.scripts, c)
}
