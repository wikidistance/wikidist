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

// requests per minute
const rateLimit = 1500

type Crawler struct {
	nWorkers   int
	startTitle string
	prefix     string

	queue          chan string
	results        chan db.Article
	notifyDequeued chan struct{}
	canMakeRequest chan struct{}
	seen           *cache.Cache
	database       db.DB
}

func NewCrawler(nWorkers int, prefix string, startTitle string, database db.DB) *Crawler {
	c := Crawler{}

	c.database = database
	c.canMakeRequest = make(chan struct{}, 10)

	c.nWorkers = nWorkers
	c.startTitle = startTitle
	c.prefix = prefix

	c.notifyDequeued = make(chan struct{}, 1)
	c.queue = make(chan string, queueSizeFactor*nWorkers)
	c.results = make(chan db.Article, resultQueueSizeFactor*nWorkers)
	c.seen = cache.New(30*time.Minute, 30*time.Minute)

	return &c
}

// until we find a better solution
func (c *Crawler) AlwaysRefill() {
	for {
		c.notifyDequeued <- struct{}{}
	}
}

func (c *Crawler) Start() {

	for i := 1; i <= c.nWorkers; i++ {
		go c.addWorker()
	}

	for i := 1; i <= c.nWorkers; i++ {
		go c.addRegisterer()
	}

	go c.metrics()
	go c.rateLimit()
	go c.refillQueue()
	go c.AlwaysRefill()

	c.queue <- c.startTitle
}

func (c *Crawler) metrics() {
	for range time.Tick(10 * time.Second) {
		metrics.Statsd.Gauge("wikidist.crawler.queue.length", float64(len(c.queue)), nil, 1)
		metrics.Statsd.Gauge("wikidist.crawler.results.length", float64(len(c.results)), nil, 1)
	}
}

func (c *Crawler) refillQueue() {
	for {
		<-c.notifyDequeued
		if len(c.queue) <= refillFactor*c.nWorkers {
			titles, err := c.database.NextsToVisit(queueSizeFactor * c.nWorkers)
			if err != nil {
				log.Println(err)
			}

			var newTitles int64
			metrics.Statsd.Count("wikidist.queue.returned_titles", int64(len(titles)), nil, 1)
			for _, title := range titles {
				if _, ok := c.seen.Get(title); ok {
					continue
				}
				if len(c.queue) >= cap(c.queue) {
					break
				}
				c.seen.Set(title, struct{}{}, cache.DefaultExpiration)
				newTitles++
				c.queue <- title
			}
			metrics.Statsd.Count("wikidist.queue.new_titles", newTitles, nil, 1)
		}
	}
}

func (c *Crawler) addRegisterer() {
	for {
		result := <-c.results
		resultCopy := result

		log.Println("Registering", result.Title, len(result.LinkedArticles), result.Missing)
		start := time.Now()
		c.database.AddVisited(&resultCopy)
		log.Println("Completed AddVisited in", time.Since(start))
		log.Println("Registered", result.Title)
		metrics.Statsd.Count("wikidist.articles.registered", 1, nil, 1)
	}
}

func (c *Crawler) addWorker() {
	for {
		title := <-c.queue

		// non-blocking write to channel
		select {
		case c.notifyDequeued <- struct{}{}:
		default:
		}

		if title == "" {
			continue
		}

		<-c.canMakeRequest

		log.Println("Fetching", title)

		article, err := CrawlArticle(title, c.prefix)

		if err != nil {
			log.Println(err)

			// try putting the title back in the queue
			select {
			case c.queue <- title:
			default:
			}
			continue
		}

		c.results <- article
		metrics.Statsd.Count("wikidist.articles.fetched", 1, nil, 1)
	}
}

func (c *Crawler) rateLimit() {
	for range time.Tick(time.Minute / rateLimit) {
		c.canMakeRequest <- struct{}{}
	}
}
