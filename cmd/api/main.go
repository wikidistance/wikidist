package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/wikidistance/wikidist/pkg/api"
	"github.com/wikidistance/wikidist/pkg/db"
)

func main() {
	// Get filename from args
	filename := "../../config.json"
	if len(os.Args) == 2 {
		filename = os.Args[1]
	}

	// Get config from file
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal("Couldn't open config file")
	}

	decoder := json.NewDecoder(file)
	var config db.Config
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatal("Couldn't parse config file")
	}

	d, err := db.NewDGraph(config)
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
