package api

import (
	"github.com/wikidistance/wikidist/pkg/db"
	"log"
)

func PageSearch(s string) []db.WebPage {
	dgraph, err := db.NewDGraph()

	if err != nil {
		log.Println(err)
	}

	res, err := dgraph.SearchArticleByTitle(s)

	if err != nil {
		log.Printf("DB error: %s", err)
	}

	return res
}
