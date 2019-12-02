package db

import (
	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"google.golang.org/grpc"
)

// DGraph is a connection to a dgraph instance
type DGraph struct {
	client *dgo.Dgraph
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

	// Setup DB
	// TODO

	return &dgraph, nil
}

func AddVisited(article *Article) error {
	// TODO
	return nil
}

func NextToVisit() (string, error) {
	// TODO
	return "", nil
}
