package crawler

import (
	"fmt"
	"github.com/wikidistance/wikidist/pkg/db"
	"sync"
)

type Crawler struct {
	nWorkers int
	startUrl string

	results chan db.Article
	toSee   map[string]struct{}
	l       sync.Mutex

	seen  map[string]struct{}
	graph map[string]db.Article
}

func NewCrawler(nWorkers int, startUrl string) *Crawler {
	c := Crawler{}

	c.nWorkers = nWorkers
	c.startUrl = startUrl

	c.results = make(chan db.Article, nWorkers)
	c.seen = make(map[string]struct{})
	c.toSee = make(map[string]struct{})
	c.graph = make(map[string]db.Article)

	return &c
}

func (c *Crawler) Run() {
	nQueued := 1
	c.toSee[c.startUrl] = struct{}{}
	c.seen[c.startUrl] = struct{}{}

	for i := 1; i <= c.nWorkers; i++ {
		go c.addWorker()
	}

	for nCrawled := 0; nQueued > nCrawled; nCrawled++ {
		result := <-c.results
		fmt.Println("got result", result.Title, len(result.LinkedArticles))

		c.graph[result.URL] = result
		for _, link := range result.LinkedArticles {
			if _, ok := c.seen[link]; !ok {
				nQueued++

				c.l.Lock()
				c.toSee[link] = struct{}{}
				c.l.Unlock()

				c.seen[link] = struct{}{}
			}
		}

		fmt.Println(nQueued, "queued,", nCrawled, "crawled")
	}
}

func (c *Crawler) addWorker() {
	for {
		var url string
		c.l.Lock()
		for link := range c.toSee {
			url = link
			break
		}
		delete(c.toSee, url)
		c.l.Unlock()

		if url == "" {
			continue
		}

		fmt.Println("getting", url)
		c.results <- CrawlArticle(url)
	}
}
