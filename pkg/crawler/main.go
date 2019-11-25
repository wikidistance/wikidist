package crawler

import (
	"net/http"
	s "strings"

	"golang.org/x/net/html"
)

func getPageLinks(url string) (pageLinks []string) {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	z := html.NewTokenizer(resp.Body)

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
							pageLinks = append(pageLinks, a.Val)
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
