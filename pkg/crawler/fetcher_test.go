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
			<a href="http://location_outside_wikipedia.com">Outside</a>
			<p><a id="test" href="/wiki/Article_about_something_else">Another article</a></p> 
		</body>
	</html>`))
	expected := []string{"/wiki/Article_about_something_else"}

	assert.Equal(t, extractLinks(fakeBody), expected, "Only the second link should be extracted")
}
