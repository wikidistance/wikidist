package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/wikidistance/wikidist/pkg/crawler"
	"github.com/wikidistance/wikidist/pkg/db"
)

func main() {

	client, _ := db.NewDGraph()
	c := crawler.NewCrawler(40, "/wiki/Alan_Turing", client)

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	c.Run()
}
