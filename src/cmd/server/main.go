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
  http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){
    w.Header().Set("Content-Type", "text/plain")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("It's working!!"))
  })
  port := os.Getenv("PORT")
  fmt.Println("Listen in port", port)
  http.ListenAndServe(port, nil)
}
