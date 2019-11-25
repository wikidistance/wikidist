package crawler

import (
	"io"
	"net/http"
	s "strings"

	"golang.org/x/net/html"
)

func GetPageLinks(url string) []string {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	return extractLinks(resp.Body)
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
