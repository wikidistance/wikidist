package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/wikidistance/wikidist/pkg/crawler"
	"github.com/wikidistance/wikidist/pkg/db"
)

func main() {

	args := os.Args[1:]

	log.Println(args)

	if len(args) != 3 && len(args) != 4 {
		log.Println("Usage: crawler <prefix> <startTitle> <nWorkers> <config>")
		return
	}

	nWorkers, err := strconv.Atoi(args[2])

	if err != nil {
		log.Println("nWorkers should be an integer")
		return
	}

	// Get filename from args
	filename := "../../config.json"
	if len(args) == 4 {
		filename = args[3]
	}

	// Get config from file
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal("Couldn't open config file")
	}

	var config db.Config
	err = json.Unmarshal(file, &config)
	if err != nil {
		log.Fatal("Couldn't parse config file")
	}

	client, err := db.NewDGraph(config)
	if err != nil {
		log.Fatal("Couldn't connect to DGraph")
	}

	c := crawler.NewCrawler(nWorkers, args[0], args[1], client)

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	c.Start()

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	<-done
	log.Println("exiting")
}
