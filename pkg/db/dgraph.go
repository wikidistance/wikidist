package db

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"github.com/wikidistance/wikidist/pkg/metrics"
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
	offset      int

	createGroup singleflight.Group
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
	dgraph.offset = 0

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

	ctx := context.Background()

	//get the uids of the linked articles
	uids, err := dg.fetchArticles(ctx, article.LinkedArticles)
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

	// remove the linked articles not to create duplicates
	article.LinkedArticles = nil
	uid, err := dg.getOrCreate(ctx, article)
	if err != nil {
		return err
	}

	// now that we know for sure the uid of the article, let's mutate it again
	article.UID = uid
	article.DType = []string{"Article"}
	article.LinkedArticles = linkedArticles
	article.LastCrawled = time.Now().Format("2006-01-02T15:04:05Z")

	pb, err := json.Marshal(article)
	if err != nil {
		return err
	}
	mu := &api.Mutation{
		SetJson:   pb,
		CommitNow: true,
	}
	log.Println("linking", article.URL)
	_, err = dg.client.NewTxn().Mutate(ctx, mu)
	log.Println("linked", article.URL)

	return err
}

// getOrCreate returns the uid of the article based on the URL whether created or already existing
func (dg *DGraph) getOrCreate(ctx context.Context, article *Article) (string, error) {
	txn := dg.client.NewTxn()
	defer txn.Discard(ctx)
	uid, err := dg.getOrCreateWithTxn(ctx, txn, article)
	if err != nil {
		return uid, err
	}

	txn.Commit(ctx)
	return uid, err
}

func (dg *DGraph) getOrCreateWithTxn(ctx context.Context, txn *dgo.Txn, article *Article) (string, error) {
	uid, err, _ := dg.createGroup.Do(article.URL, func() (interface{}, error) {

		uid, err := dg.queryArticle(ctx, article)
		if err != nil {
			return "", err
		}
		if uid != "" {
			return uid, err
		}

		article.UID = "_:article"
		article.DType = []string{"Article"}
		article.LastCrawled = dummyDate
		pb, err := json.Marshal(article)
		if err != nil {
			return "", err
		}

		mu := &api.Mutation{
			SetJson: pb,
		}

		log.Println("adding", article.URL)
		resp, err := txn.Mutate(ctx, mu)
		if err != nil {
			return "", err
		}
		uid = resp.Uids["article"]
		log.Println("added", article.URL, uid)

		return uid, nil
	})

	return uid.(string), err
}

func (dg *DGraph) fetchArticles(ctx context.Context, articles []Article) ([]string, error) {
	uids := make([]string, 0, len(articles))

	txn := dg.client.NewTxn()
	for _, article := range articles {

		uid, err := dg.getOrCreateWithTxn(ctx, txn, &article)
		if err != nil {
			return nil, err
		}
		uids = append(uids, uid)
	}

	err := txn.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return uids, nil

}

func (dg *DGraph) queryArticle(ctx context.Context, article *Article) (string, error) {

	txn := dg.client.NewReadOnlyTxn().BestEffort()
	defer txn.Discard(ctx)

	q := `
	query Get($url: string) {
		get(func: eq(url, $url)) {
			uid,
			url,
			title
		}
	}
	`

	// check cache
	if uid, ok := dg.cacheLookup(article.URL); ok {
		metrics.Statsd.Count("wikidist.uidcache.hit", 1, nil, 1)
		return uid, nil
	}

	metrics.Statsd.Count("wikidist.uidcache.miss", 1, nil, 1)

	r, err := dg.query(ctx, txn, q, map[string]string{"$url": article.URL})
	if err != nil {
		return "", err
	}

	if len(r["get"]) > 0 {
		if len(r["get"]) > 1 {
			panic(fmt.Sprintf("There shouldn't ever be more than one node with same URL: %s\n", article.URL))
		}

		uid := r["get"][0].UID

		// save in cache

		dg.cacheSave(article.URL, uid)
	}

	return "", nil
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
	ctx := context.TODO()

	txn := dg.client.NewReadOnlyTxn().BestEffort()

	var query = fmt.Sprintf(`
	{
		nodes(func: eq(last_crawled, "%s"), first: %d, offset: %d) {
			uid
			url
			title
		}
	}
	`, dummyDate, count, dg.offset*count)

	dg.offset++
	dg.offset %= 10

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
		if node.Title != "" {
			fmt.Println("NextToVisit returned an already crawled article:", node.URL)
		}
	}

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
