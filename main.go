package main

import (
	"flag"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

var (
	base         = flag.String("d", os.Getenv("HOME"), "directory to serve")
	addr         = flag.String("http", ":6060", "addr:port for server")
	templateName = "dir.gohtml"
)

func main() {
	flag.Parse()

	r := gin.Default()
	r.GET("/*path", handlePath)

	r.Run(*addr)
}

func handlePath(c *gin.Context) {
	tmpl := template.Must(template.New(templateName).Funcs(template.FuncMap{
		"ext":  filepath.Ext,
		"dir":  filepath.Dir,
		"join": path.Join,
	}).ParseFiles(templateName))
	path := path.Join(*base, c.Param("path"))
	p, err := os.Stat(path)
	if err != nil {
		c.Redirect(http.StatusFound, "/")
		return
	}

	if !p.IsDir() {
		c.File(path)
		return
	}

	files, err := ioutil.ReadDir(path)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal error while reading directory"})
		return
	}
	tmpl.Execute(c.Writer, struct {
		Base  string
		Path  string
		Files []os.FileInfo
	}{
		Path:  c.Param("path"),
		Files: files,
	})

	c.Writer.Flush()

	return
}
