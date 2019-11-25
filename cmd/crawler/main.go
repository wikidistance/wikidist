package main

import (
	"fmt"

	"github.com/wikidistance/wikidist/pkg/crawler"
)

func main() {
	fmt.Println(crawler.GetPageLinks("https://fr.wikipedia.org/wiki/Vilebrequin_(moteur)"))
}
