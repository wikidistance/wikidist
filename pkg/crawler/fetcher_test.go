package crawler

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type RoundTripper struct{}

func (rt *RoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	switch req.Method {
	case "HEAD":
		return &http.Response{Request: req}, nil
	default:
		panic(fmt.Errorf("mock http client does not implement %s requests", req.Method))
	}
}

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

	fakeClient := http.DefaultClient
	fakeClient.Transport = &RoundTripper{}

	title, links := parsePage(fakeClient, fakeBody)
	expectedLinks := []string{"/wiki/Article_about_something_else"}
	expectedTitle := "Title of the page"

	assert.Equal(t, links, expectedLinks, "Only the second link should be extracted")
	assert.Equal(t, title, expectedTitle, "Page title should be extracted")
}
