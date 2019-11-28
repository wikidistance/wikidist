package crawler

import (
	"io"
	"net/http"
	s "strings"

	"golang.org/x/net/html"
)

func GetPageLinks(url string) []string {
	prefix := "https://en.wikipedia.org"
	resp, err := http.Get(prefix + url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	return removeDuplicates(extractLinks(resp.Body))
}

func extractLinks(pageBody io.ReadCloser) (links []string) {
	z := html.NewTokenizer(pageBody)

	for {
		tt := z.Next()
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
