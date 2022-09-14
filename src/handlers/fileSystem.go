package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"
)

func sendJSON(w http.ResponseWriter, r *http.Request, b interface{}) {
	json.NewEncoder(w).Encode(b)
}

func toHTTPError(err error) (string, int) {
	if errors.Is(err, fs.ErrNotExist) {
		return "404 page not found", http.StatusNotFound
	}
	if errors.Is(err, fs.ErrPermission) {
		return "403 Forbidden", http.StatusForbidden
	}
	// Default:
	return "500 Internal Server Error", http.StatusInternalServerError
}

type dirEntry struct {
	Name string `json:"name"`
}
type fileEntry struct {
	Name    string    `json:"name"`
	ModTime time.Time `json:"modTime"`
	Size    int64     `json:"size"`
	Type    string    `json:"type"`
}

type ServeDirectoryResponse struct {
	Directories []dirEntry  `json:"directories"`
	Files       []fileEntry `json:"files"`
}

func serverDirectory(w http.ResponseWriter, r *http.Request, dir *os.File) {
	dirs, err := dir.Readdir(128)
	if err != nil && !errors.Is(err, io.EOF) {
		msg, code := toHTTPError(err)
		http.Error(w, msg, code)
		return
	}
	var body = ServeDirectoryResponse{}
	for _, d := range dirs {
    if d.IsDir(){
      body.Directories = append(body.Directories, dirEntry{ Name: d.Name()})
    }else{
      body.Files = append(body.Files, fileEntry{
        Name: d.Name(),
        ModTime: d.ModTime(),
        Size: d.Size(),
        Type: mime.TypeByExtension(filepath.Ext(d.Name())),
      })
    }
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(body)
}

func FileStatsHandler(root string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
    // cors
		w.Header().Set("access-control-allow-origin", "*")

		upath := path.Clean(r.URL.Path)
		dirPath := path.Join(root, upath)
		dir, err := os.Open(dirPath)
		if err != nil {
			msg, code := toHTTPError(err)
			http.Error(w, msg, code)
			return
		}
		defer dir.Close()
		stats, err := dir.Stat()
		if err != nil {
			msg, code := toHTTPError(err)
			http.Error(w, msg, code)
			return
		}
		if stats.IsDir() {
			serverDirectory(w, r, dir)
			return
		}
		action := r.URL.Query().Get("action")
		if action == "download" {
			w.Header().Add("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filepath.Base(dirPath)))
		}
		http.ServeContent(w, r, stats.Name(), stats.ModTime(), dir)
	}
}
