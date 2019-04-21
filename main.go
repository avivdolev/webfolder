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
	templateFile = "dir.gohtml"
	tmpl         *template.Template
)

func main() {
	flag.Parse()

	tmpl = makeTemplate()

	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/dir")
		return
	})
	r.GET("/dir/*path", handlePath)
	r.POST("/upload/*path", handleUpload)

	r.Run(*addr)
}

func makeTemplate() *template.Template {
	return template.Must(template.New(templateFile).Funcs(template.FuncMap{
		"ext":  filepath.Ext,
		"Dir":  filepath.Dir,
		"join": path.Join,
	}).ParseFiles(templateFile))
}

func handlePath(c *gin.Context) {
	fpath := filepath.Join(*base, c.Param("path"))
	p, err := os.Stat(fpath)
	if err != nil {
		c.Redirect(http.StatusFound, "/dir")
		return
	}

	if !p.IsDir() {
		c.File(fpath)
		return
	}

	files, err := ioutil.ReadDir(fpath)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal error while reading directory"})
		return
	}
	tmpl.Execute(c.Writer, struct {
		Path  string
		Files []os.FileInfo
	}{
		Path:  c.Param("path"),
		Files: files,
	})

	c.Writer.Flush()

	return
}

func handleUpload(c *gin.Context) {
	fpath := filepath.Join(*base, c.Param("path"))
	form, err := c.MultipartForm()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "bad form"})
		return
	}

	files := form.File["files"]

	for _, file := range files {
		fname := filepath.Join(fpath, file.Filename)
		if err := c.SaveUploadedFile(file, fname); err != nil {
			continue
		}
	}
	c.Redirect(http.StatusFound, path.Join("/dir", c.Param("path")))
}
