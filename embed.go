package stock

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"path"

	"github.com/gin-contrib/static"
)

//go:embed all:web/dist
var StaticDir embed.FS

type embedFileSystem struct {
	http.FileSystem
}

func (e embedFileSystem) Exists(prefix string, path string) bool {
	_, err := e.Open(path)
	// if err != nil {
	//      return false
	// }
	return err == nil
}

func EmbedFolder() static.ServeFileSystem {
	fsys, err := fs.Sub(StaticDir, "web/dist")
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	return embedFileSystem{
		FileSystem: http.FS(fsys),
	}
}

type EmbedFileHTTP struct {
}

func (f *EmbedFileHTTP) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fPath := path.Join("web/dist", "index.html")

	embedFs, err := StaticDir.Open(fPath)
	if err != nil {
		fmt.Println(err)
	}
	fStat, err := embedFs.Stat()
	if err != nil {
		fmt.Println(err)
	}
	http.ServeContent(w, r, fStat.Name(), fStat.ModTime(), embedFs.(io.ReadSeeker))
}
