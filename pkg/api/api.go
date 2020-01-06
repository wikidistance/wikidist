package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/wikidistance/wikidist/pkg/db"
)

type DGraph db.DGraph

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
