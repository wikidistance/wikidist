package main

import (
	"fmt"
	"github.com/wikidistance/wikidist/pkg/crawler"
)

type Result struct {
	url   string
	links []string
}

func main() {
	nWorkers := 3

	queue := make(chan string, 10000)
	defer close(queue)

	results := make(chan Result, nWorkers)

	seen := make(map[string]struct{})
	graph := make(map[string][]string)

	startUrl := "/wiki/Amauroclopius"
	queue <- startUrl
	seen[startUrl] = struct{}{}

	for i := 1; i <= nWorkers; i++ {
		go worker(queue, results)
	}

	for nCrawled := 0; nCrawled <= 100; nCrawled++ {
		result := <-results
		fmt.Println("got result", result.url, len(result.links))

		graph[result.url] = result.links
		for _, link := range result.links {
			if _, already_seen := seen[link]; !already_seen {
				queue <- link
				seen[link] = struct{}{}
			}
		}
	}
}

func worker(queue <-chan string, results chan<- Result) {
	for url := range queue {
		fmt.Println("getting", url)
		links := crawler.GetPageLinks(url)
		results <- Result{url, links}
	}
}
