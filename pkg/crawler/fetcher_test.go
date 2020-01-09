package crawler

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFetcher(t *testing.T) {
	fakeBody := ioutil.NopCloser(strings.NewReader(`<!DOCTYPE html>
	<html>
		<head></head>
		<body>
			<h1 id="firstHeading">Title of the page</h1>
			<a href="http://location_outside_wikipedia.com">Outside</a>
			<a href="/wiki/Sample_Page">Outside</a>
			<p><a id="test" href="/wiki/Article_about_something_else">Another article</a></p> 
		</body>
	</html>`))
	title, links := parsePage("/wiki/Sample_Page", fakeBody)
	expectedLinks := []string{"/wiki/Article_about_something_else"}
	expectedTitle := "Title of the page"

	assert.Equal(t, links, expectedLinks, "Only the second link should be extracted")
	assert.Equal(t, title, expectedTitle, "Page title should be extracted")
}
