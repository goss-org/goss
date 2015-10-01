package system

type Gossfile struct {
	path string
}

func (g *Gossfile) Path() string {
	return g.path
}

func NewGossfile(path string, system *System) *Gossfile {
	return &Gossfile{path: path}
}
