package crawler

import (
	"io"
	"net/http"
	s "strings"

	"golang.org/x/net/html"
)

func GetPageLinks(url string) Article {
	prefix := "https://en.wikipedia.org"
	resp, err := http.Get(prefix + url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	title, links := parsePage(resp.Body)
	return Article{url, title, removeDuplicates(links)}
}

func parsePage(pageBody io.ReadCloser) (title string, links []string) {
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
							links = append(links, a.Val)
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
	return s.HasPrefix(link, "/wiki/") && !s.Contains(link, ":")
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
