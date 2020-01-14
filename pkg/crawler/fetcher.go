package crawler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/wikidistance/wikidist/pkg/db"
	"github.com/wikidistance/wikidist/pkg/metrics"
)

// CrawlArticle : Crawls an article given its title
func CrawlArticle(title string, prefix string) (db.Article, error) {
	baseURL := "https://" + prefix + ".wikipedia.org/w/api.php?format=json&action=query&prop=links|description&pllimit=500&plnamespace=0"

	// TODO: Pagination logic
	resp, err := http.Get(baseURL + "&titles=" + url.QueryEscape(title))
	if err != nil {
		log.Printf("Request failed for article %s: %w", title, err)
		return db.Article{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		log.Println("Request failed for article", title, ", status", resp.StatusCode)
	}

	body, _ := ioutil.ReadAll(resp.Body)

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	missing, description, links, err := parseResponse(result)

	if err != nil {
		log.Println("Error while fetching article", title, ":", err)
		return db.Article{}, err
	}

	if missing {
		log.Println("Article", title, "is missing")
		metrics.Statsd.Count("wikidist.crawler.articles.missing", 1, nil, 1)
	}

	linkedArticles := make([]db.Article, 0, len(links))
	for _, link := range links {
		linkedArticles = append(linkedArticles, db.Article{
			Title: link,
		})
	}

	return db.Article{
		Title:          title,
		Description:    description,
		Missing:        missing,
		LinkedArticles: linkedArticles,
	}, nil
}

func parseResponse(response map[string]interface{}) (bool, string, []string, error) {
	query := response["query"].(map[string]interface{})
	titles := make([]string, 0)
	for _, value := range (query["pages"]).(map[string]interface{}) {
		page := value.(map[string]interface{})

		// handle when page is missing
		if _, ok := page["missing"]; ok {
			return true, "", []string{}, nil
		}

		description := ""
		if desc, ok := page["description"]; ok {
			description = desc.(string)
		}

		if _, ok := page["links"]; !ok {
			return false, description, []string{}, nil
		}

		links := page["links"].([]interface{})
		for _, value := range links {
			link := value.(map[string]interface{})
			switch link["title"].(type) {
			case string:
				titles = append(titles, link["title"].(string))
			default:
				return true, "", []string{}, fmt.Errorf("Incorrect title in answer")
			}
		}
		return false, description, titles, nil
	}

	return true, "", []string{}, fmt.Errorf("No page in answer")
}
