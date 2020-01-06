package main

import (
	"github.com/wikidistance/wikidist/pkg/crawler"
	"github.com/wikidistance/wikidist/pkg/db"
)

func main() {

	client, _ := db.NewDGraph()
	// fmt.Println(client.NextToVisit())
	c := crawler.NewCrawler(10, "/wiki/Luni", client)

	c.Run()
}
