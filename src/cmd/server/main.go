package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/JesusJMM/cloud_file_system/src/handlers"

	"github.com/joho/godotenv"
)

func main() {
  err := godotenv.Load()
  if err != nil {
    log.Fatal("Error loading .env")
  }

  storage_dir := os.Getenv("STORAGE_DIR")

  fs := http.FileServer(http.Dir(storage_dir))
  http.Handle("/raw/", http.StripPrefix("/raw/", fs))
  http.Handle("/browse/", http.StripPrefix("/browse/", handlers.FileStatsHandler(storage_dir)))

  port := os.Getenv("PORT")
  fmt.Println("Listen in port", port)
  http.ListenAndServe(port, nil)
}
