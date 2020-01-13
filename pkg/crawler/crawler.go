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
const refillFactor = queueSizeFactor / 2

type Crawler struct {
	nWorkers int
	startURL string
	prefix   string

	queue          chan string
	results        chan db.Article
	notifyDequeued chan bool
	seen           *cache.Cache
	database       db.DB
}

func NewCrawler(nWorkers int, prefix string, startURL string, database db.DB) *Crawler {
	c := Crawler{}

	c.database = database

	c.nWorkers = nWorkers
	c.startURL = startURL
	c.prefix = prefix

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
	go c.refillQueue()

	c.queue <- c.startURL
}

func (c *Crawler) metrics() {
	for range time.Tick(10 * time.Second) {
		metrics.Statsd.Gauge("wikidist.crawler.queue.length", float64(len(c.queue)), nil, 1)
		metrics.Statsd.Gauge("wikidist.crawler.results.length", float64(len(c.results)), nil, 1)
	}
}

func (c *Crawler) refillQueue() {
	for <-c.notifyDequeued {
		if len(c.queue) <= refillFactor*c.nWorkers {
			urls, err := c.database.NextsToVisit(queueSizeFactor * c.nWorkers)
			if err != nil {
				log.Println(err)
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
	}
}

func (c *Crawler) addRegisterer() {
	for {
		result := <-c.results
		resultCopy := result

		log.Println("Registering", result.URL, result.Title, len(result.LinkedArticles))
		c.database.AddVisited(&resultCopy)
		metrics.Statsd.Count("wikidist.articles.registered", 1, nil, 1)
	}
}

func (c *Crawler) addWorker() {
	for {
		url := <-c.queue

		// non-blocking write to channel
		select {
		case c.notifyDequeued <- true:
		default:
		}

		if url == "" {
			continue
		}

		article, err := CrawlArticle(url, c.prefix)
		if err != nil {
			log.Println(err)

			// try putting the url back in the queue
			select {
			case c.queue <- url:
			default:
			}
			continue
		}

		c.results <- article
		metrics.Statsd.Count("wikidist.articles.fetched", 1, nil, 1)
	}
}
