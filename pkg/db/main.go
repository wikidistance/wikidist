package db

// Article contains information about a MediaWiki article
type Article struct {
	UID            string    `json:"uid,omitempty"`
	Title          string    `json:"title,omitempty"`
	Description    string    `json:"description,omitempty"`
	Missing        bool      `json:"missing,omitempty"`
	LinkedArticles []Article `json:"linked_articles,omitempty"`
	LastCrawled    string    `json:"last_crawled,omitempty"`
	DType          []string  `json:"D_type,omitempty"`
	PageID         int       `json:"page_id,omitempty"`
}

// DB is the interface to interact with a database
type DB interface {
	// AddVisited writes the visited article and its edges with other articles.
	// It should be called after each article has been visited.
	AddVisited(*Article) error

	// NextsToVisit returns a list of Titles at random from the list of articles that have yet
	// to be visited. If there is nothing to visit, this function blocks
	// indefinitely until there is one.
	NextsToVisit(count int) ([]string, error)
}
