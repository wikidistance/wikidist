package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/wikidistance/wikidist/pkg/api"
	"github.com/wikidistance/wikidist/pkg/db"
)

func main() {
	d, err := db.NewDGraph()
	dg := (*api.DGraph)(d)

	if err != nil {
		log.Fatal("Couldn't connect to DGraph")
	}

	http.HandleFunc("/", api.DefaultHandler)
	http.HandleFunc("/shortest", dg.ShortestPathHandler)
	http.HandleFunc("/search", dg.PageSearchHandler)
	http.HandleFunc("/search-uid", dg.UidSearchHandler)

	fmt.Println("API is running on port 8081")

	log.Fatal(http.ListenAndServe(":8081", nil))
}
