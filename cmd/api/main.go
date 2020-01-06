package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/wikidistance/wikidist/pkg/api"
)

func main() {
	http.HandleFunc("/", api.DefaultHandler)
	http.HandleFunc("/shortest", api.ShortestPathHandler)
	fmt.Println("APi is running on port 8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
