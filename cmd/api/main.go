package main

import (
	"log"
	"net/http"

	"github.com/wikidistance/wikidist/pkg/api"
)

func main() {
	http.HandleFunc("/", api.DefaultHandler)
	http.HandleFunc("/search", api.PageSearchHandler)

	log.Fatal(http.ListenAndServe(":8081", nil))
}
