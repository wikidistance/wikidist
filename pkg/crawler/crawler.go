package crawler

import (
	"log"
	"time"

	"github.com/patrickmn/go-cache"

	"github.com/wikidistance/wikidist/pkg/db"
	"github.com/wikidistance/wikidist/pkg/metrics"
)

// ratio to nWorkers
const queueSizeFactor = 200
const resultQueueSizeFactor = 2
const refillFactor = queueSizeFactor * 10 / 8

type Crawler struct {
	nWorkers int
	startURL string

	queue    chan string
	results  chan db.Article
	seen     *cache.Cache
	database db.DB
}

func NewCrawler(nWorkers int, startURL string, database db.DB) *Crawler {
	c := Crawler{}

	c.database = database

	c.nWorkers = nWorkers
	c.startURL = startURL

	c.queue = make(chan string, queueSizeFactor*nWorkers)
	c.results = make(chan db.Article, resultQueueSizeFactor*nWorkers)
	c.seen = cache.New(5*time.Minute, 5*time.Minute)

	return &c
}

func (c *Crawler) Run() {

	for i := 1; i <= c.nWorkers; i++ {
		go c.addWorker()
	}

	for i := 1; i <= c.nWorkers; i++ {
		go c.addRegisterer()
	}

	go c.metrics()

	c.queue <- c.startURL

	for {
		err := c.refillQueue()
		if err != nil {
			log.Println(err)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func (c *Crawler) metrics() {
	for {
		metrics.Statsd.Gauge("wikidist.crawler.queue.length", float64(len(c.queue)), nil, 1)
		metrics.Statsd.Gauge("wikidist.crawler.results.length", float64(len(c.results)), nil, 1)
		time.Sleep(10 * time.Second)
	}
}

func (c *Crawler) refillQueue() error {
	if len(c.queue) <= refillFactor*c.nWorkers {
		urls, err := c.database.NextsToVisit(queueSizeFactor * c.nWorkers)
		if err != nil {
			return err
		}

		var newURLs int64
		metrics.Statsd.Count("wikidist.queue.returned_urls", int64(len(urls)), nil, 1)
		for _, url := range urls {
			if _, ok := c.seen.Get(url); ok {
				continue
			}
			if len(c.queue) >= cap(c.queue) {
				break
			}
			c.seen.Set(url, struct{}{}, cache.DefaultExpiration)
			newURLs++
			c.queue <- url
		}
		metrics.Statsd.Count("wikidist.queue.new_urls", newURLs, nil, 1)

	}

	return nil
}

func (c *Crawler) addRegisterer() {
	for {
		result := <-c.results
		resultCopy := result

		log.Println("registering", result.URL)
		c.database.AddVisited(&resultCopy)
		metrics.Statsd.Count("wikidist.articles.registered", 1, nil, 1)
	}
}

func (c *Crawler) addWorker() {
	for {
		url := <-c.queue

		if url == "" {
			continue
		}

		c.results <- CrawlArticle(url)
		metrics.Statsd.Count("wikidist.articles.fetched", 1, nil, 1)
	}
}
