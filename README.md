![](https://github.com/wikidistance/wikidist/workflows/Test/badge.svg)

### Starting dgraph locally

```sh
docker-compose up -d
```

### Starting the server

```sh
cd cmd/api
go build
./api
```

### Running unit tests for the fetcher

`fetcher.go` is unit tested with 100% code coverage.

```sh
cd pkg/crawler
go test -coverprofile=coverage.out
go tool cover -html=coverage.out
```
