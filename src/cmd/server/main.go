package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
  err := godotenv.Load()
  if err != nil {
    log.Fatal("Error loading .env")
  }

  fs := http.FileServer(http.Dir(os.Getenv("STORAGE_DIR")))
  http.Handle("/raw/", http.StripPrefix("/raw/", fs))

  port := os.Getenv("PORT")
  fmt.Println("Listen in port", port)
  http.ListenAndServe(port, nil)
}
