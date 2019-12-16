package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/wikidistance/wikidist/pkg/db"
)

type Search struct {
	Search string `json:"search"`
}

func DefaultHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World")
}

func PageSearchHandler(w http.ResponseWriter, r *http.Request) {

	var search Search
	var res []db.WebPage

	w.Header().Set("Content-type", "application/json;charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	reqBody, err := ioutil.ReadAll(r.Body)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Please enter a valid search string")
	}

	json.Unmarshal(reqBody, &search)

	res = PageSearch(search.Search)

	json.NewEncoder(w).Encode(res)
}
