package crawler

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	s "strings"
	"time"

	"github.com/wikidistance/wikidist/pkg/db"
	"github.com/wikidistance/wikidist/pkg/metrics"
	"golang.org/x/net/html"
)

func CrawlArticle(url string, prefix string) (db.Article, error) {
	prefix = "https://" + prefix + ".wikipedia.org"

	start := time.Now()
	resp, err := http.Get(prefix + url)
	elapsed := time.Since(start)
	metrics.Statsd.Gauge("wikidist.fetcher.time", float64(elapsed.Milliseconds()), nil, 1)

	if err != nil {
		return db.Article{}, fmt.Errorf("failed to fetch article %s: %w", url, err)
	}

	defer resp.Body.Close()

	title, links := parsePage(url, resp.Body)

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
	}, nil
}

func parsePage(url string, pageBody io.ReadCloser) (title string, links []string) {
	var buf bytes.Buffer
	tee := io.TeeReader(pageBody, &buf)

	z := html.NewTokenizer(tee)

	titleIsNext := false
	done := false

	for !done {
		tt := z.Next()
		if titleIsNext {
			titleIsNext = false
			title = string(z.Raw())
		}
		switch {
		case tt == html.ErrorToken:
			done = true
		case tt == html.StartTagToken:
			t := z.Token()
			if t.Data == "a" {
				for _, a := range t.Attr {
					if a.Key != "href" {
						continue
					}
					// Handle links to section: /path/to/doc#section
					link := s.SplitN(a.Val, "#", 2)[0]
					if isLinkToArticle(link) && url != link {
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

	if title == "" {
		metrics.Statsd.Count("wikidist.article.notitle", 1, nil, 1)

		fmt.Printf(buf.String())
	}

	return
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
