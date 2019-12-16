package db

// Article contains information about a MediaWiki article
type Article struct {
	UID            string    `json:"uid,omitempty"`
	URL            string    `json:"url,omitempty"`
	Title          string    `json:"title,omitempty"`
	LinkedArticles []Article `json:"linked_articles,omitempty"`
	DType          []string
}

// DB is the interface to interact with a database
type DB interface {
	// AddVisited writes the visited article and its edges with other articles.
	// It should be called after each article has been visited.
	AddVisited(*Article) error

	// NextToVisit returns a URL at random from the list of URLs that have yet.
	// to be visited. If there is nothing to visit, this function blocks
	// indefinitely until there is one.
	NextToVisit() (string, error)
}
