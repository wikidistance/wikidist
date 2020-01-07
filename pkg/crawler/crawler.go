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
	seen     map[string]struct{}
	database db.DB
}

func NewCrawler(nWorkers int, startURL string, database db.DB) *Crawler {
	c := Crawler{}

	c.database = database

	c.nWorkers = nWorkers
	c.startURL = startURL

	c.queue = make(chan string, 100*nWorkers)
	c.results = make(chan db.Article, 2*nWorkers)
	c.seen = make(map[string]struct{})

	return &c
}

func (c *Crawler) Run() {

	for i := 1; i <= c.nWorkers; i++ {
		go c.addWorker()
	}

	for i := 1; i <= c.nWorkers; i++ {
		go c.addRegisterer()
	}

	c.queue <- c.startURL

	for {
		c.refillQueue()
		time.Sleep(10 * time.Millisecond)
	}
}

func (c *Crawler) refillQueue() {
	if len(c.queue) <= 50*c.nWorkers {
		urls, err := c.database.NextsToVisit(50 * c.nWorkers)
		if err != nil {
			panic(err)
		}

		for _, url := range urls {
			if _, ok := c.seen[url]; ok {
				continue
			}
			c.seen[url] = struct{}{}
			fmt.Println("queuing", url)
			c.queue <- url
		}
	}
}

func (c *Crawler) addRegisterer() {
	for {
		result := <-c.results
		resultCopy := result

		fmt.Println("registering", result.URL)
		c.database.AddVisited(&resultCopy)
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
