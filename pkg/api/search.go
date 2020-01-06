package api

import (
	"github.com/wikidistance/wikidist/pkg/db"
	"log"
)

func PageSearch(s string, depth int) []db.Article {
	dgraph, err := db.NewDGraph()

	if err != nil {
		log.Println(err)
	}

	res, err := dgraph.SearchArticleByTitle(s, depth)

	if err != nil {
		log.Printf("DB error: %s", err)
	}

	return res
}
