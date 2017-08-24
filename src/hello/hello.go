package main

import (
	"io"
	"net/http"
	"log"
)

func HelloServer(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "hello")
}

func main() {
	http.HandleFunc("/hello", HelloServer)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("listene", err)
	}
}