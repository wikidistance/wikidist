package crawler

import (
	"io"
	"log"
	"net/http"
	s "strings"

	"github.com/wikidistance/wikidist/pkg/db"
	"golang.org/x/net/html"
)

func CrawlArticle(url string) db.Article {
	prefix := "https://fr.wikipedia.org"
	resp, err := http.Get(prefix + url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	title, links := parsePage(http.DefaultClient, resp.Body)

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

func parsePage(client *http.Client, pageBody io.ReadCloser) (title string, links []string) {
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
					if a.Key == "href" {
						if isLinkToArticle(a.Val) {
							// Handle links to section: /path/to/doc#section
							link := s.SplitN(a.Val, "#", 2)[0]

							// Do a head request and follow redirects
							// to ensure we have the actual article URL
							res, err := client.Head(link)
							if err != nil {
								log.Printf("failed to fetch %s: %s", link, err)
							}
							links = append(links, res.Request.URL.String())
						}
						break
					}
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
