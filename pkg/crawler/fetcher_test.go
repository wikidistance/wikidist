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
			<p><a id="test" href="/wiki/Article_about_something_else">Another article</a></p> 
		</body>
	</html>`))
	title, links := parsePage(fakeBody)
	expected_links := []string{"/wiki/Article_about_something_else"}
	expected_title := "Title of the page"

	assert.Equal(t, links, expected_links, "Only the second link should be extracted")
	assert.Equal(t, title, expected_title, "Page title should be extracted")
}
