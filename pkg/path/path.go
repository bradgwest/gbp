package path

import "strings"

// PathSep is the path separator
const PathSep = "/"

// Path models paths like /the/url/path/{id}
type Path struct {
	Path string
	ID   string
}

// New creates a new path
func New(p string) *Path {
	var id string
	p = strings.Trim(p, PathSep)
	s := strings.Split(p, PathSep)
	if len(s) > 1 {
		id = s[len(s)-1]
		p = strings.Join(s[:len(s)-1], PathSep)
	}
	return &Path{Path: p, ID: id}
}

// HasID returns true if the path has an id
func (p *Path) HasID() bool {
	return len(p.ID) > 0
}
