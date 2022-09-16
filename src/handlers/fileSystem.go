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
	"strconv"
	"time"
)

const (
	MAX_UPLOAD_SIZE = 20 << 20
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
		if d.IsDir() {
			body.Directories = append(body.Directories, dirEntry{Name: d.Name()})
		} else {
			body.Files = append(body.Files, fileEntry{
				Name:    d.Name(),
				ModTime: d.ModTime(),
				Size:    d.Size(),
				Type:    mime.TypeByExtension(filepath.Ext(d.Name())),
			})
		}
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(body)
}

func handleGet(w http.ResponseWriter, r *http.Request, dirPath string) {
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

func handleUploadFile(w http.ResponseWriter, r *http.Request, dirPath string) {
	if err := r.ParseMultipartForm(MAX_UPLOAD_SIZE); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
    return
	}
	files := r.MultipartForm.File["files"]
	for _, fh := range files {
		fmt.Printf("Processing File: %s", fh.Filename)
		if fh.Size > MAX_UPLOAD_SIZE {
			http.Error(w, "The uploaded file is too big. Please use a file less than 1MB in size", http.StatusBadRequest)
			return
		}
		file, err := fh.Open()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer file.Close()

		f, err := os.Create(path.Join(dirPath, fmt.Sprintf("%s-%s", strconv.FormatInt(time.Now().UnixNano(), 32), fh.Filename)))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer f.Close()
		_, err = io.Copy(f, file)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Printf("Uploaded file: %s", fh.Filename)
	}
	w.WriteHeader(http.StatusCreated)
}

func handleDeleteFile(w http.ResponseWriter, r *http.Request, dirPath string){
  file, err := os.Open(dirPath)
  fmt.Printf("Deleting %s", dirPath)
  if err != nil {
    msg, code := toHTTPError(err)
    http.Error(w, msg, code)
    return
  }
  stats, err := file.Stat()
  if err != nil {
    msg, code := toHTTPError(err)
    http.Error(w, msg, code)
    return
  }
  if stats.IsDir(){
    err := os.RemoveAll(dirPath)
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }
    w.WriteHeader(200)
    return
  }
  os.Remove(dirPath)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  w.WriteHeader(200)
  fmt.Printf("%s Deleted successfully", dirPath)
}

func handleCreateFolder(w http.ResponseWriter, r *http.Request, dirPath string){
  fmt.Println("creating folder ", dirPath)
  err := os.Mkdir(dirPath, 0755)
  if err != nil {
    if os.IsExist(err) {
      http.Error(w, "This folder already exist", 400)
      return
    }
    msg, code := toHTTPError(err)
    http.Error(w, msg, code)
    return
  }
  w.WriteHeader(201)
}

func FileStatsHandler(root string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
    // cors
    w.Header().Set("access-control-allow-origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "POST, GET, DELETE")

    upath := path.Clean(r.URL.Path)
    dirPath := path.Join(root, upath)

		switch r.Method {
		case http.MethodGet:
			handleGet(w, r, dirPath)
		case http.MethodPost:
      resourceType := r.URL.Query().Get("type")
      if resourceType == "folder" {
        handleCreateFolder(w,r,dirPath)
        return
      }
			handleUploadFile(w, r, dirPath)
    case http.MethodDelete:
      if upath == "" {
        w.Write([]byte("This route is protected."))
        w.WriteHeader(http.StatusForbidden)
        return
      }
      handleDeleteFile(w, r, dirPath)
		}
	}
}
