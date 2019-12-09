package main

import (
	"fmt"

	"github.com/wikidistance/wikidist/pkg/db"
)

func main() {
	//c := crawler.NewCrawler(10, "/wiki/Main_Page")

	client, _ := db.NewDGraph()

	err := client.AddVisited(&db.Article{
		Title: "titlea",
		URL:   "urla",
		LinkedArticles: []db.Article{
			{
				URL: "urlb",
			},
		},
	})

	fmt.Println(err)

	//c.Run()
}
