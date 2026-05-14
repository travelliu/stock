package stock

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gin-contrib/static"
)

//go:embed all:web/dist
var StaticDir embed.FS

type embedFS struct{ http.FileSystem }

func (e embedFS) Exists(prefix, filepath string) bool {
	if _, err := e.Open(filepath); err != nil {
		return false
	}
	return true
}

func EmbedFolder() static.ServeFileSystem {
	sub, _ := fs.Sub(StaticDir, "web/dist")
	return embedFS{http.FS(sub)}
}
