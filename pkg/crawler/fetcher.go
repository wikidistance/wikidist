package crawler

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	s "strings"
	"time"

	"github.com/wikidistance/wikidist/pkg/db"
	"github.com/wikidistance/wikidist/pkg/metrics"
	"golang.org/x/net/html"
)

const prefix = "https://fr.wikipedia.org"

func CrawlArticle(url string) db.Article {
	start := time.Now()
	resp, err := http.Get(prefix + url)
	elapsed := time.Since(start)
	metrics.Statsd.Gauge("wikidist.fetcher.time", float64(elapsed.Milliseconds()), nil, 1)

	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	title, links := parsePage(http.DefaultClient, url, resp.Body)

	dedupedLinks := removeDuplicates(links)

	// build neighbour Articles
	linkedArticles := make([]db.Article, 0, len(dedupedLinks))
	for _, link := range dedupedLinks {
		neighbour := db.Article{URL: link}
		linkedArticles = append(linkedArticles, neighbour)

	}

	return db.Article{
		URL:            url,
		Title:          title,
		LinkedArticles: linkedArticles,
	}
}

func parsePage(client *http.Client, url string, pageBody io.ReadCloser) (title string, links []string) {
	z := html.NewTokenizer(pageBody)

	titleIsNext := false

	for {
		tt := z.Next()
		if titleIsNext {
			titleIsNext = false
			title = string(z.Raw())
		}
		switch {
		case tt == html.ErrorToken:
			return
		case tt == html.StartTagToken:
			t := z.Token()
			if t.Data == "a" {
				for _, a := range t.Attr {
					if a.Key != "href" {
						continue
					}
					// Handle links to section: /path/to/doc#section
					link := s.SplitN(a.Val, "#", 2)[0]
					if !isLinkToArticle(link) {
						continue
					}

					// Do a head request and follow redirects
					// to ensure we have the actual article URL
					res, err := client.Head(prefix + link)
					if err != nil {
						log.Printf("failed to fetch %s: %s", link, err)
						continue
					}

					link = strings.TrimPrefix(prefix, res.Request.URL.String())
					fmt.Println("after head", link)
					if url != link {
						links = append(links, link)
					}
					break

				}
			}
			if t.Data == "h1" {
				for _, a := range t.Attr {
					if a.Key == "id" && a.Val == "firstHeading" {
						titleIsNext = true
					}
				}
			}
		}
	}
}

func isLinkToArticle(link string) bool {
	return s.HasPrefix(link, "/wiki/") && !s.Contains(link, ":") && link != "/wiki/Main_Page" && link != "/wiki/Pagina_maestra"
}

func removeDuplicates(links []string) (dedupedLinks []string) {
	hashTable := make(map[string]bool)
	for _, link := range links {
		hashTable[link] = true
	}

	for link := range hashTable {
		dedupedLinks = append(dedupedLinks, link)
	}

	return
}
