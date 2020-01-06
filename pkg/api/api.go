package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/wikidistance/wikidist/pkg/db"
)

type DGraph db.DGraph

type Search struct {
	Search string `json:"search"`
	Depth  int    `json:"depth"`
}

func DefaultHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World")
}

func (dg *DGraph) ShortestPathHandler(w http.ResponseWriter, r *http.Request) {
	dg2 := (*db.DGraph)(dg)
	from, ok := r.URL.Query()["from"]
	fmt.Fprintf(w, "from[0] = %s, from = %s", from[0], from)
	if !ok || len(from[0]) < 1 {
		fmt.Fprint(w, "Need a from argument")
		return
	}
	to, ok := r.URL.Query()["to"]
	if !ok || len(to[0]) < 1 {
		fmt.Fprint(w, "Need a to argument")
		return
	}

	res, err := dg2.ShortestPath(from[0], to[0])
	if err != nil {
		log.Printf("DB error: %s", err)
	}

	json.NewEncoder(w).Encode(res)
}

func (dg *DGraph) PageSearchHandler(w http.ResponseWriter, r *http.Request) {

	var search Search
	var res []db.Article

	w.Header().Set("Content-type", "application/json;charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	reqBody, err := ioutil.ReadAll(r.Body)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Please enter a valid search string")
	}

	json.Unmarshal(reqBody, &search)

	res = PageSearch(search.Search, search.Depth)

	json.NewEncoder(w).Encode(res)
}
