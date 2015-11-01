package main

import (
	"io"
	"net/http"
	"os"
)

func index(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "index")
}

func main() {
	port := os.Getenv("PORT")

	http.HandleFunc("/", index)
	http.ListenAndServe(":"+port, nil)
}
