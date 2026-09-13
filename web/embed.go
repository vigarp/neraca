package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var distFS embed.FS

// GetFS mengembalikan filesystem sub-tree dari folder "dist"
func GetFS() (fs.FS, error) {
	return fs.Sub(distFS, "dist")
}
