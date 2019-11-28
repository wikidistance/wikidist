package db

type Article struct {
	URL            string
	Title          string
	LinkedArticles []string
}

type DB interface {
	// AddVisited writes the visited article and its edges with other articles.
	// It should be called after each article has been visited.
	AddVisited(*Article) error

	// NextToVisit returns a URL at random from the list of URLs that have yet.
	// to be visited. If there is nothing to visit, this function blocks
	// indefinitely until there is one.
	NextToVisit() (string, error)
}
