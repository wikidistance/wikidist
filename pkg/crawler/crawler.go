package crawler

import (
	"fmt"
	"time"

	"github.com/wikidistance/wikidist/pkg/db"
)

type Crawler struct {
	nWorkers int
	startURL string

	queue    chan string
	results  chan db.Article
	database db.DB
}

func NewCrawler(nWorkers int, startURL string, database db.DB) *Crawler {
	c := Crawler{}

	c.database = database

	c.nWorkers = nWorkers
	c.startURL = startURL

	c.queue = make(chan string, 10*nWorkers)
	c.results = make(chan db.Article, 100*nWorkers)

	return &c
}

func (c *Crawler) Run() {

	seen := make(map[string]struct{})
	for i := 1; i <= c.nWorkers; i++ {
		go c.addWorker()
	}

	nCrawled := 0
	c.queue <- c.startURL

	for {
		// fill queue
		if len(c.queue) <= c.nWorkers {
			urls, err := c.database.NextsToVisit(9 * c.nWorkers)
			if err != nil {
				panic(err)
			}

			for _, url := range urls {
				if _, ok := seen[url]; ok {
					continue
				}
				seen[url] = struct{}{}
				fmt.Println("queuing", url)
				c.queue <- url
			}
		}

		time.Sleep(time.Millisecond)

		// save results
		if len(c.results) > 0 {
			result := <-c.results

			fmt.Println("got result", result.Title, len(result.LinkedArticles))
			resultCopy := result

			c.database.AddVisited(&resultCopy)
			fmt.Println("registered", result.Title)
			nCrawled++

			fmt.Println(nCrawled, "crawled")
		}
	}
}

func (c *Crawler) addWorker() {
	for {
		url := <-c.queue

		if url == "" {
			continue
		}

		fmt.Println("getting", url)
		c.results <- CrawlArticle(url)
	}
}
