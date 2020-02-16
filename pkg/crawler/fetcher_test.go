package crawler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/wikidistance/wikidist/pkg/db"
)

type mockHTTPGet struct {
	resp      *http.Response
	shouldErr bool
}

func (mock mockHTTPGet) Get(rawURL string) (*http.Response, error) {
	if mock.shouldErr {
		return nil, fmt.Errorf("mock Get() was instructed to return an error")
	}
	return mock.resp, nil
}

type testCase struct {
	// filepath to the test case
	path string

	// mock http.Get()
	mockHTTPGet httpGetter

	// whether the call to CrawlArticle() should return an error
	shouldErr bool

	// expected db.Article
	expectedArticle db.Article
}

// loadTestCases loads test cases from JSON files
func loadTestCases(basepath string) ([]*testCase, error) {
	var testCases []*testCase
	files, err := ioutil.ReadDir(basepath)
	if err != nil {
		return nil, fmt.Errorf("failed to list files in directory %s: %s", basepath, err)
	}

	for _, f := range files {
		if f.IsDir() {
			path := filepath.Join(basepath, f.Name())

			tc := testCase{path: path}
			err = tc.load()
			if err != nil {
				return nil, fmt.Errorf("failed to load test case %s: %s", path, err)
			}
			testCases = append(testCases, &tc)
		}
	}
	return testCases, nil
}

type testMetadata struct {
	// whether the call to CrawlArticle should return an error
	TestShouldErr bool

	// whether our mock Get() func should return an error
	MockShouldErr bool

	// which HTTP status our mock Get() func should return
	HTTPStatusCode int
}

func (tc *testCase) load() error {
	// Load test metadata
	metadataPath := filepath.Join(tc.path, "metadata.json")
	b, err := ioutil.ReadFile(metadataPath)
	if err != nil {
		return err
	}
	var m testMetadata
	err = json.Unmarshal(b, &m)
	if err != nil {
		return fmt.Errorf("failed to parse JSON from file %s: %s", metadataPath, err)
	}

	tc.shouldErr = m.TestShouldErr

	// Load expected db.Article
	articlePath := filepath.Join(tc.path, "article.json")
	if _, err := os.Stat(articlePath); err == nil {
		// file exists
		b, err = ioutil.ReadFile(articlePath)
		if err != nil {
			return err
		}
		err = json.Unmarshal(b, &tc.expectedArticle)
		if err != nil {
			return fmt.Errorf("failed to parse JSON from file %s: %s", articlePath, err)
		}
	}

	// Load mock response body
	var body []byte
	bodyPath := filepath.Join(tc.path, "body.json")
	if _, err := os.Stat(bodyPath); err == nil {
		// file exists
		body, err = ioutil.ReadFile(bodyPath)
		if err != nil {
			return err
		}
	}

	tc.mockHTTPGet = &mockHTTPGet{
		resp: &http.Response{
			Status:        http.StatusText(m.HTTPStatusCode),
			StatusCode:    m.HTTPStatusCode,
			Proto:         "HTTP/1.1",
			ProtoMajor:    1,
			ProtoMinor:    1,
			Body:          ioutil.NopCloser(bytes.NewBuffer(body)),
			ContentLength: int64(len(body)),
			Header:        make(http.Header, 0),
		},
		shouldErr: m.MockShouldErr,
	}
	return nil
}

const PathToTestCases = "./test_cases"
const TestTitle = "TEST_TITLE"
const TestPrefix = "TEST_PREFIX"

func TestCrawlArticle(t *testing.T) {
	testCases, err := loadTestCases(PathToTestCases)
	if err != nil {
		t.Fatalf("failed to load test cases: %s", err)
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			article, err := CrawlArticle(TestTitle, TestPrefix, tc.mockHTTPGet)

			if tc.shouldErr {
				if err == nil {
					t.Fatalf("expected test to return an error, got nil instead")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpectedly returned error: %s", err)
				}

				if !reflect.DeepEqual(tc.expectedArticle, article) {
					t.Fatalf("db.Article mismatch: expected %v, got %v", tc.expectedArticle, article)
				}
			}
		})
	}
}
