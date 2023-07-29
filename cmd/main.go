package main

import (
	"compress/gzip"
	_ "embed"
	"encoding/base64"
	"fmt"
	"html/template"
	"math/rand"
	"strconv"

	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	picturePath = "./pictures"
)

//go:embed index.html
var indexHTML string

var indexTpl = template.Must(template.New("").Parse(indexHTML))

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	if err := updatePictures(); err != nil {
		log.Println(err)
		return
	}

	mux.HandleFunc("/", Handle)
	mux.HandleFunc("/update", Update)
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("OK")) })
	log.Println("listen and server at :9288")
	log.Println(suffix)
	if err := http.ListenAndServe(":9288", mux); err != nil {
		log.Println(err)
		return
	}
}

var (
	mux      = http.NewServeMux()
	pictures struct {
		mu       sync.Mutex
		pictures []string
	}
	suffix = getPictureSuffix()
)

func getPictureSuffix() map[string]bool {
	v := make(map[string]bool)
	s := os.Getenv("PICTURE_SUFFIX")
	if s == "" {
		s = "png,jpg,jpeg"
	}
	for _, i := range strings.Split(s, ",") {
		i = strings.TrimSpace(i)
		if !strings.HasSuffix(i, ".") {
			i = "." + i
		}
		v[i] = true
	}
	return v
}

func updatePictures() error {
	files, err := FilePathWalkDir(picturePath)
	if err != nil {
		log.Println(err)
		return err
	}
	pictures.mu.Lock()
	pictures.pictures = files
	pictures.mu.Unlock()
	log.Printf("find %d pictures\n", len(files))
	return nil
}

func FilePathWalkDir(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, info os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if info.Type().IsRegular() {
			if suffix[filepath.Ext(path)] {
				files = append(files, path)
			}
		}
		return nil
	})
	return files, err
}

type Data struct {
	Current int
	Next    int
	Size    int
	Picture string
	Name    string
}

func Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	err := updatePictures()
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	w.WriteHeader(200)

	w.Write([]byte(fmt.Sprintf(`<pre>error: %v</pre><div><a href="/">Return</a></div>`, err)))
}

func Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	r.ParseForm()
	idx, err := strconv.Atoi(r.Form.Get("idx"))
	pictures.mu.Lock()
	defer pictures.mu.Unlock()

	var d Data
	if len(pictures.pictures) != 0 {
		if err != nil || idx < 0 || idx >= len(pictures.pictures) {
			idx = rand.Intn(len(pictures.pictures))
		}
		name := pictures.pictures[idx]
		data, err := os.ReadFile(name)
		if err != nil {
			log.Println("read picture fail: ", name, err)
		}
		d.Picture = base64.StdEncoding.EncodeToString(data)
		d.Name = filepath.Base(name)
	}
	d.Current = idx
	d.Next = idx + 1
	d.Size = len(pictures.pictures)

	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	w.Header().Set("Content-Encoding", "gzip")
	w.WriteHeader(200)
	gz := gzip.NewWriter(w)
	defer gz.Close()

	if err := indexTpl.Execute(gz, &d); err != nil {
		log.Println(err)
	}
}
