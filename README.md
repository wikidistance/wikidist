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

### Frontend setup

#### Setting env

To change API url in frontend you can edit `frontend/.env`

By default the adress is set to the VM hosted by CentraleSupélec but you can only access it from CS network.
If you want to test locally, you can put the local API host adress and port (by default `localhost:8081`)

``` js
VUE_APP_API_URL="http://localhost:8081"
```

#### Serve files

You can serve the frontend either whith

```
cd frontend
yarn serve
```

or by building files and serving staticly the built html files

```
yarn build
```