package main

import (
	"fmt"
	"sync"

	"github.com/wikidistance/wikidist/pkg/crawler"
)

var (
	toSee = make(map[string]struct{})
	l     = sync.Mutex{}
)

type Result struct {
	url   string
	links []string
}

func main() {
	nWorkers := 10

	queue := make(chan string, 10000)
	defer close(queue)

	results := make(chan Result, nWorkers)

	seen := make(map[string]struct{})
	graph := make(map[string][]string)

	nQueued := 1

	startUrl := "/wiki/Osasco"
	queue <- startUrl
	seen[startUrl] = struct{}{}

	for i := 1; i <= nWorkers; i++ {
		go worker(queue, results)
	}

	nCrawled := 0

	for ; nCrawled <= 10000 && nQueued > nCrawled; nCrawled++ {
		result := <-results
		fmt.Println("got result", result.url, len(result.links))

		graph[result.url] = result.links
		for _, link := range result.links {
			if _, ok := seen[link]; !ok {
				nQueued++
				select {
				case queue <- link:
				default:
					l.Lock()
					toSee[link] = struct{}{}
					l.Unlock()
				}
				seen[link] = struct{}{}
			}
		}

		fmt.Println(nQueued, "queued,", nCrawled, "crawled")
	}
}

func worker(queue <-chan string, results chan<- Result) {
	for {
		var url string
		select {
		case url = <-queue:
		default:
			l.Lock()
			for link := range toSee {
				url = link
				break
			}
			delete(toSee, url)
			l.Unlock()
		}
		if url == "" {
			continue
		}

		fmt.Println("getting", url)
		links := crawler.GetPageLinks(url)
		results <- Result{url, links}
	}
}
