package api

import (
	"github.com/wikidistance/wikidist/pkg/db"
	"log"
)

func (dg *DGraph) PageSearch(s string, depth int) []db.Article {
	dg2 := (*db.DGraph)(dg)

	res, err := dg2.SearchArticleByTitle(s, depth)

	if err != nil {
		log.Printf("DB error: %s", err)
	}

	return res
}
