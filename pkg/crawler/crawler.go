package crawler

import "fmt"

import "sync"

type Result struct {
	url   string
	links []string
}

type Crawler struct {
	nWorkers int
	startUrl string

	queue   chan string
	results chan Result
	toSee   map[string]struct{}
	l       sync.Mutex

	seen  map[string]struct{}
	graph map[string][]string
}

func NewCrawler(nWorkers int, startUrl string) *Crawler {
	c := Crawler{}

	c.nWorkers = nWorkers
	c.startUrl = startUrl

	c.queue = make(chan string, 10000)
	c.results = make(chan Result, nWorkers)
	c.seen = make(map[string]struct{})
	c.toSee = make(map[string]struct{})
	c.graph = make(map[string][]string)

	return &c
}

func (c *Crawler) Run() {
	nQueued := 1
	c.queue <- c.startUrl
	c.seen[c.startUrl] = struct{}{}

	for i := 1; i <= c.nWorkers; i++ {
		go c.addWorker()
	}

	for nCrawled := 0; nQueued > nCrawled; nCrawled++ {
		result := <-c.results
		fmt.Println("got result", result.url, len(result.links))

		c.graph[result.url] = result.links
		for _, link := range result.links {
			if _, ok := c.seen[link]; !ok {
				nQueued++
				select {
				case c.queue <- link:
				default:
					c.l.Lock()
					c.toSee[link] = struct{}{}
					c.l.Unlock()
				}
				c.seen[link] = struct{}{}
			}
		}

		fmt.Println(nQueued, "queued,", nCrawled, "crawled")
	}
}

func (c *Crawler) addWorker() {
	for {
		var url string
		select {
		case url = <-c.queue:
		default:
			c.l.Lock()
			for link := range c.toSee {
				url = link
				break
			}
			delete(c.toSee, url)
			c.l.Unlock()
		}
		if url == "" {
			continue
		}

		fmt.Println("getting", url)
		links := GetPageLinks(url)
		c.results <- Result{url, links}
	}
}
