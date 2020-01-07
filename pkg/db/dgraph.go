package db

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"google.golang.org/grpc"
)

var dummyDate = time.Date(2000, time.January, 0, 0, 0, 0, 0, time.UTC).Format("2006-01-02T15:04:05Z")

// DGraph is a connection to a dgraph instance
type DGraph struct {
	client      *dgo.Dgraph
	uidCache    map[string]string
	cacheLock   sync.Mutex
	cacheHits   int
	cacheMisses int
}

// NewDGraph returns a new *DGraph
func NewDGraph() (*DGraph, error) {
	// Dial a gRPC connection. The address to dial to can be configured when
	// setting up the dgraph cluster.
	d, err := grpc.Dial("localhost:9080", grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	dgraph := DGraph{
		client: dgo.NewDgraphClient(api.NewDgraphClient(d)),
	}

	dgraph.uidCache = make(map[string]string)
	dgraph.cacheLock = sync.Mutex{}

	op := &api.Operation{
		Schema: `type Article {
			title: string
			url: string
			linked_articles: [Article]
			last_crawled: dateTime
		}

		title: string @index(term) @lang .
		url: string @index(hash) @lang .
		last_crawled: dateTime @index(hour) .

		`,
	}

	err = dgraph.client.Alter(context.Background(), op)

	return &dgraph, err
}

func (dg *DGraph) cacheLookup(url string) (uid string, ok bool) {
	dg.cacheLock.Lock()
	defer dg.cacheLock.Unlock()

	uid, ok = dg.uidCache[url]

	return uid, ok
}

func (dg *DGraph) cacheSave(url string, uid string) {
	dg.cacheLock.Lock()
	defer dg.cacheLock.Unlock()

	dg.uidCache[url] = uid
}

func (dg *DGraph) AddVisited(article *Article) error {
	start := time.Now()
	ctx := context.Background()

	//get the uids of the linked articles
	uids, err := dg.getOrCreate(ctx, article.LinkedArticles)
	if err != nil {
		return err
	}

	// add the uids
	linkedArticles := make([]Article, 0, len(article.LinkedArticles))
	for _, uid := range uids {
		linkedArticles = append(linkedArticles, Article{
			UID: uid,
		})
	}
	article.LinkedArticles = linkedArticles

	// query whether the article already exist
	resp, err := dg.queryArticles(ctx, []Article{*article})
	if err != nil {
		return err
	}

	article.UID = "_:article"
	article.DType = []string{"Article"}

	// use the real uid if the article is already created
	if len(resp) > 0 {
		article.UID = resp[0].UID
	}

	article.LastCrawled = time.Now().Format("2006-01-02T15:04:05Z")

	// update the article with all the new links
	pb, err := json.Marshal(article)
	if err != nil {
		return err
	}
	mu := &api.Mutation{
		SetJson:   pb,
		CommitNow: true,
	}
	_, err = dg.client.NewTxn().Mutate(ctx, mu)

	log.Printf("AddVisited: processed article %s in %v\n", article.Title, time.Since(start))
	return err
}

func (dg *DGraph) getOrCreate(ctx context.Context, articles []Article) ([]string, error) {
	start := time.Now()
	uids := make([]string, 0, len(articles))

	// get the already existing articles
	existingArticles, err := dg.queryArticles(ctx, articles)
	if err != nil {
		return nil, err
	}

	existing := make(map[string]struct{})
	for _, article := range existingArticles {
		existing[article.URL] = struct{}{}
		uids = append(uids, article.UID)
	}

	// create the non-existing articles
	txn := dg.client.NewTxn()
	for _, article := range articles {
		if _, ok := existing[article.URL]; ok {
			continue
		}

		article.UID = "_:article"
		article.DType = []string{"Article"}
		article.LastCrawled = dummyDate
		pb, err := json.Marshal(article)
		if err != nil {
			return nil, err
		}

		mu := &api.Mutation{
			SetJson: pb,
		}
		resp, err := txn.Mutate(ctx, mu)
		if err != nil {
			return nil, err
		}

		uids = append(uids, resp.Uids["article"])
	}
	err = txn.Commit(ctx)
	if err != nil {
		return nil, err
	}

	log.Printf("getOrCreate: processed %d articles in %v\n", len(articles), time.Since(start))
	return uids, nil

}

func (dg *DGraph) queryArticles(ctx context.Context, articles []Article) ([]Article, error) {
	start := time.Now()
	txn := dg.client.NewReadOnlyTxn().BestEffort()
	defer txn.Discard(ctx)

	resp := make([]Article, 0, len(articles))
	q := `
	query Get($url: string) {
		get(func: eq(url, $url)) {
			uid,
			url
		}
	}
	`

	cacheHits, cacheMisses := 0, 0

	for _, article := range articles {
		// check cache
		if uid, ok := dg.cacheLookup(article.URL); ok {
			cacheHits++
			resp = append(resp, Article{
				UID: uid,
				URL: article.URL,
			})
			continue
		}

		cacheMisses++

		r, err := dg.query(ctx, txn, q, map[string]string{"$url": article.URL})
		if err != nil {
			return nil, err
		}
		if len(r["get"]) > 0 {
			resp = append(resp, r["get"][0])

			// save in cache

			dg.cacheSave(article.URL, r["get"][0].UID)
		}

	}

	if cacheHits+cacheMisses > 0 {
		log.Printf("Cache hits: %d, misses: %d, hit ratio: %d%%\n", cacheHits, cacheMisses, 100*cacheHits/(cacheHits+cacheMisses))
	}
	log.Printf("queryArticles: queried %d articles in %v\n", len(articles), time.Since(start))
	return resp, nil
}

func (dg *DGraph) query(ctx context.Context, txn *dgo.Txn, q string, vars map[string]string) (map[string][]Article, error) {
	resp, err := txn.QueryWithVars(ctx, q, vars)
	if err != nil {
		return nil, err
	}

	r := make(map[string][]Article)
	err = json.Unmarshal(resp.GetJson(), &r)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (dg *DGraph) NextsToVisit(count int) ([]string, error) {
	start := time.Now()
	ctx := context.TODO()

	txn := dg.client.NewReadOnlyTxn().BestEffort()

	var query = fmt.Sprintf(`
	{
		nodes(func: eq(last_crawled, "%s"), first: %d) {
			uid
			url
		}
	}
	`, dummyDate, count)

	resp, err := txn.Query(ctx, query)
	if err != nil {
		fmt.Println(err)
	}

	var decode struct {
		Nodes []Article
	}

	if err := json.Unmarshal(resp.GetJson(), &decode); err != nil {
		fmt.Println(err)
	}

	urls := make([]string, 0)

	for _, node := range decode.Nodes {
		urls = append(urls, node.URL)
	}

	log.Printf("NextToVisit: finished in %v\n", time.Since(start))
	return urls, nil
}

func (dg *DGraph) ShortestPath(from string, to string) ([]Article, error) {
	q := fmt.Sprintf(`
	{
		path as shortest(from: %s, to: %s) {
			linked_articles
		   }
		path(func: uid(path)) {
			uid,
			title,
			url
		}
	}

	`, from, to)
	resp, err := dg.client.NewTxn().Query(context.Background(), q)

	if err != nil {
		return nil, err
	}
	result := make(map[string][]Article, 0)
	println("resp", resp.String())
	json.Unmarshal(resp.GetJson(), &result)
	return result["path"], nil

}
