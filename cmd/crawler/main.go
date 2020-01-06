package main

import (
	"github.com/wikidistance/wikidist/pkg/crawler"
	"github.com/wikidistance/wikidist/pkg/db"
)

func main() {

	client, _ := db.NewDGraph()
	c := crawler.NewCrawler(10, "/wiki/Alan_Turing", client)

	c.Run()
}
