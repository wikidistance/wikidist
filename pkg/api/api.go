package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/wikidistance/wikidist/pkg/db"
)

var dg *db.DGraph

func init() {
	dg, _ = db.NewDGraph()
}
func DefaultHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World")
}

func ShortestPathHandler(w http.ResponseWriter, r *http.Request) {
	from, ok := r.URL.Query()["from"]
	if !ok || len(from[0]) < 1 {
		fmt.Fprint(w, "Need a from argument")
		return
	}
	to, ok := r.URL.Query()["to"]
	if !ok || len(to[0]) < 1 {
		fmt.Fprint(w, "Need a to argument")
		return
	}
	res, err := dg.ShortestPath(from[0], to[0])
	if err != nil {
		log.Printf("DB error: %s", err)
	}

	json.NewEncoder(w).Encode(res)

}
