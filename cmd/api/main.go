package main

import (
	"log"
	"net/http"

	"github.com/wikidistance/wikidist/pkg/api"
)

func main() {
	http.HandleFunc("/", api.DefaultHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
