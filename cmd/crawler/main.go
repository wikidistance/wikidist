package main

import "github.com/wikidistance/wikidist/pkg/crawler"

func main() {
	c := crawler.NewCrawler(10, "/wiki/Main_Page")

	c.Run()
}
