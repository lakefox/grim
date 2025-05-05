package grim

type Scripts struct {
	scripts []Script
}

type Script struct {
	Call func(*Node)
}

func (s *Scripts) Run(n *Node) {
	for _, v := range s.scripts {
		v.Call(n)
	}
}

func (s *Scripts) Add(c Script) {
	s.scripts = append(s.scripts, c)
}
