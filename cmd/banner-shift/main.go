package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func testHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<h1>Test OK!</h1>")
}

func main() {
	// TODO: add config

	// TODO: add logger

	// TODO: storage: postresql

	// TODO: router: go-chi

	// TODO: run server
	// fmt.Println("test")

	_ = chi.NewRouter()

	http.HandleFunc("/", testHandler)
	http.ListenAndServe(":8080", nil)
}
