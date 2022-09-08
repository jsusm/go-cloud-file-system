package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
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
	Name    string    `json:"name"`
	IsDir   bool      `json:"isDir"`
	ModTime time.Time `json:"modTime"`
	Size    int64     `json:"size"`
}

func serverDirectory(w http.ResponseWriter, r *http.Request, dir *os.File) {
	dirs, err := dir.Readdir(128)
	if err != nil && !errors.Is(err, io.EOF) {
		msg, code := toHTTPError(err)
		http.Error(w, msg, code)
    return
	}
	var body = []dirEntry{}
	for _, d := range dirs {
		body = append(body, dirEntry{d.Name(), d.IsDir(), d.ModTime(), d.Size()})
	}
  w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(body)
}

func FileStatsHandler(root string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("access-control-allow-origin", "*")
		upath := path.Clean(r.URL.Path) 
		dir, err := os.Open(path.Join(root, upath))
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
		http.Redirect(w, r, path.Join("/raw/", upath), http.StatusMovedPermanently)
	}
}
