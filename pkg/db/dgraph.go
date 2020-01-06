package db

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"google.golang.org/grpc"
)

// DGraph is a connection to a dgraph instance
type DGraph struct {
	client *dgo.Dgraph
}

type WebPage struct {
	Uid            string    `json:"uid"`
	Url            string    `json:"url"`
	Title          string    `json:"title"`
	LinkedArticles []WebPage `json:"linked_articles"`
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

	op := &api.Operation{
		Schema: `type Article {
			title: string
			url: string
			linked_articles: [Article]
		}

		title: string @index(term, trigram, fulltext) @lang .
		url: string @index(term) @lang .
		`,
	}

	err = dgraph.client.Alter(context.Background(), op)

	return &dgraph, err
}

func (dg *DGraph) AddVisited(article *Article) error {
	ctx := context.TODO()

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
	q := fmt.Sprintf(`
	{
		get(func: eq(url, "%s")) {
			uid,
			url
		}
	}
	`, article.URL)
	resp, err := dg.client.NewTxn().Query(ctx, q)
	if err != nil {
		return err
	}

	r := make(map[string][]Article)
	err = json.Unmarshal(resp.GetJson(), &r)
	if err != nil {
		return err
	}

	article.UID = "_:article"
	article.DType = []string{"Article"}

	// use the real uid if the article is already created
	if len(r["get"]) > 0 {
		article.UID = r["get"][0].UID
	}

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

	return err
}

func (dg *DGraph) getOrCreate(ctx context.Context, articles []Article) ([]string, error) {
	uids := make([]string, 0, len(articles))

	// get the already existing articles
	urls := ""
	for _, article := range articles {
		urls = fmt.Sprintf("%s %s", urls, article.URL)
	}

	q := fmt.Sprintf(`
	{
		get(func: anyofterms(url, "%s")) {
		  uid
		  url
		}
	  }

	`, urls)

	resp, err := dg.client.NewTxn().Query(ctx, q)
	if err != nil {
		return nil, err
	}

	r := make(map[string][]Article)
	err = json.Unmarshal(resp.GetJson(), &r)
	if err != nil {
		return nil, err
	}

	existing := make(map[string]struct{})
	for _, article := range r["get"] {
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

	return uids, nil

}

func (dg *DGraph) NextToVisit() (string, error) {
	// TODO
	return "", nil
}

func GenerateSearchQuery(depth int) string {
	if depth == 0 {
		return `
			title
			url
			uid
		`
	}

	return fmt.Sprintf(`
		title
		url
		uid
		linked_articles {
			%s
		}
	`, GenerateSearchQuery(depth-1))
}

func (dg *DGraph) SearchArticleByTitle(s string, depth int) ([]WebPage, error) {
	ctx := context.TODO()

	q := fmt.Sprintf(`{
		find_node_by_title(func: match(title, "%s", 2))
		{
		  %s
		}
	  }`, s, GenerateSearchQuery(depth))

	result, err := dg.client.NewTxn().Query(ctx, q)

	if err != nil {
		return nil, err
	}

	res := make(map[string][]WebPage, 0)

	json.Unmarshal(result.GetJson(), &res)

	return res["find_node_by_title"], nil
}
