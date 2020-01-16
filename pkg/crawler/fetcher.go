package crawler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/wikidistance/wikidist/pkg/db"
	"github.com/wikidistance/wikidist/pkg/metrics"
)

// CrawlArticle : Crawls an article given its title
func CrawlArticle(title string, prefix string) (db.Article, error) {
	baseURL := "https://" + prefix + ".wikipedia.org/w/api.php"

	query := url.Values{}
	query.Set("format", "json")
	query.Add("action", "query")
	query.Add("prop", "links|description")
	query.Add("pllimit", "500")
	query.Add("plnamespace", "0")
	query.Add("titles", url.QueryEscape(title))

	// TODO: Pagination logic
	resp, err := http.Get(baseURL + "?" + query.Encode())
	if err != nil {
		log.Printf("Request failed for article %s: %w", title, err)
		metrics.Statsd.Count("wikidist.requests", 1, []string{"state:hard_failure"}, 1)
		return db.Article{}, err
	}
	defer resp.Body.Close()
	metrics.Statsd.Count("wikidist.requests", 1, []string{"state:" + strconv.FormatInt(int64(resp.StatusCode), 10)}, 1)
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		log.Println("Request failed for article", title, ", status", resp.StatusCode)
		return db.Article{}, fmt.Errorf("Rate limited")
	}

	body, _ := ioutil.ReadAll(resp.Body)

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	missing, description, links, err := parseResponse(result)

	if err != nil {
		log.Println("Error while fetching article", title, ":", err)
		return db.Article{}, err
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
