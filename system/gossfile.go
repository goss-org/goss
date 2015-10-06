package system

type Gossfile struct {
	path string
}

func (g *Gossfile) Path() string {
	return g.path
}

// Stub out
func (g *Gossfile) Exists() (interface{}, error) {
	return false, nil
}

func NewGossfile(path string, system *System) Gossfile {
	return Gossfile{path: path}
}
