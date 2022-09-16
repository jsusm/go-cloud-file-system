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
    fmt.Println(".env file not found, using enviroment variables.")
  }
  storage_dir := os.Getenv("STORAGE_DIR")
  if storage_dir == "" {
    log.Panic("env variable: STORAGE_DIR, must be set")
  }

  http.Handle("/browse/", http.StripPrefix("/browse/", handlers.FileStatsHandler(storage_dir)))

  port := os.Getenv("PORT")
  fmt.Println("Listen in port", port)
  http.ListenAndServe(port, nil)
}
