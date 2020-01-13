package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"

	"github.com/wikidistance/wikidist/pkg/crawler"
	"github.com/wikidistance/wikidist/pkg/db"
)

func main() {

	args := os.Args[1:]

	fmt.Println(args)

	if len(args) != 3 {
		fmt.Println("Usage: crawler <prefix> <startUrl> <nWorkers>")
		return
	}

	nWorkers, err := strconv.Atoi(args[2])

	if err != nil {
		fmt.Println("nWorkers should be an integer")
		return
	}

	client, _ := db.NewDGraph()
	c := crawler.NewCrawler(nWorkers, args[0], args[1], client)

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	c.Run()
}
