package crawler

import (
	"fmt"

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

	c.queue = make(chan string, nWorkers*2)
	c.results = make(chan db.Article, nWorkers)

	return &c
}

func (c *Crawler) Run() {
	for i := 1; i <= c.nWorkers; i++ {
		go c.addWorker()
	}

	nCrawled := 0

	for {
		// fill queue
		if len(c.queue) <= c.nWorkers {
			url, err := c.database.NextToVisit()
			if err != nil {
				panic(err)
			}

			fmt.Println("queuing", url)
			c.queue <- url
		}

		// save results
		result := <-c.results
		fmt.Println("got result", result.Title, len(result.LinkedArticles))
		resultCopy := result

		c.database.AddVisited(&resultCopy)
		nCrawled++

		fmt.Println(nCrawled, "crawled")
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
