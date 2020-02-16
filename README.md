![](https://github.com/wikidistance/wikidist/workflows/Test/badge.svg)

# About this project

**How many links do you need to click to go from one Wikipedia article to another?**
Wikidist answers this question and tells you precisely on which links to click in order to achieve this.

You can use it as a guessing game. Or just to satisfy your curiosity :-)

# Architecture overview

Wikidist has four major components, presented in the following sections.

![Architecture overview](https://user-images.githubusercontent.com/8351433/74613849-aeac2580-5112-11ea-879a-cd3ed36edea9.png)

## Graph database

The database stores a graph in which each vertex is a page from Wikipedia, and each (oriented) edge is a link from one page to another.

For the project, we chose [`dgraph`](https://dgraph.io/), a somewhat recent but [largely rising](https://db-engines.com/en/ranking_trend/system/Dgraph) database engine because:

* using a graph database (instead of a relational one) largely facilitates the computation of a shortest path between two pages
* `dgraph` is written in Go, and thus naturally has a [great Go client API](https://godoc.org/github.com/dgraph-io/dgo)
* `dgraph`'s performance exceeds the project's requirements, even when run on a single node

## Crawler

_Written in Go, code located in [`cmd/crawler`](cmd/crawler)._

The crawler, well, crawls a given instance of Wikipedia, and stores the visited or to-be-visited pages in the database.

It takes a starting page as one of its arguments, and uses the Wikipedia API to fetch the links from that page. It then creates all the corresponding vertices in the database, marking them as 'not visited', and proceeds to visit one of them. It keeps going until there is no non-visited vertex left.

The crawler uses a multi-worker architecture: a queue is filled with yet-to-visit articles from the database, the workers consume the article in the queue and fetch the linked articles from the wikimedia API. These linked articles are then updated in the database, forming new links and new unvisited articles that will be added to the queue. This architecture allows an efficient crawling of the graph with multiple articles queried at the same time. The bottleneck here is actually the rate limiting from the API.

We assume that Wikipedia is a strongly connected graph, meaning that we will reach every page starting from any page. This assumption may not be true, but is reasonable given that the users will most likely search for "popular" pages with many links.

## API server

_Written in Go, code located in [`cmd/api`](cmd/api)._

The API server receives requests from the front-end, fetches the necessary information from the database, and responds with the corresponding data.

It has three endpoints.

### **/search**
Usage:
```shell
POST /search {"search": "<your_search>", "depth": <depth>}
```

Result:
```json
[
    {
        "title": "Title",
        "uid": "0xbeef",
        "linked_articles":[
            {...}
        ]
    },
    {...}
]
```

The search endpoint is meant to be used to find an article in the database given its title. It returns its `title`, `uid` and `linked_articles`. You can select the depth of the result (in terms of `linked_articles`). By default, the depth is 0 (i.e. there won't be any `linked_articles` in the result).

**Warning:**
This endpoint may return multiple results, as the back-end fetches all articles at a (Levenshtein) distance of 2 from the requested title. This feature can be useful to correct typing errors or to offer an autocomplete functionality.

### **/search-uid**

Usage:
```shell
POST /search-uid {"search": "<uid>", "depth": <depth>}
```

Result:
```json
[
    {
        "title": "Title",
        "uid": "0xbeef",
        "linked_articles":[
            {...}
        ]
    },
    {...}
]
```

The search-uid endpoint is meant to be used to find an article in the database given its uid. It returns its `title`, `uid` and `linked_articles`. You can select the depth of the result (in terms of `linked_articles`). By default, the depth is 0 (i.e. there won't be any `linked_articles` in the result).

This endpoint is mainly used to find the neighbors of an article.

### **/shortest**

Usage:
```shell
GET /shortest?from=<source-uid>&to=<target-uid>
```

Result:
```json
[
    {
        "title": "Source",
        "uid": "0xbeef"
    },
    {...},
    {
        "title": "Target",
        "uid": "0xdeaf"
    }
]
```

The shortest endpoint is meant to be used to find the shortest path between two articles of the database.

It returns the titles and uids of all the interim articles along the route.

## Front-end

_Code located in [`frontend`](frontend)._

TODO: technologies used, technical walkthrough...

# Installation

## Requirements

To run this project, you will need to have the following applications installed:

- docker
- docker-compose
- node + yarn
- go

Make sure that ports 5080, 6080, 8000, 8080, 8085 and 9080 are not currently in use on your machine.

The front-end will be accessible on your local port 8085.
The database console will be accessible on local port 8000.

## The easy way

The project is provided with a `docker-compose` file for an easy and fast installation. You can run the entire project minus the crawler by simply running:

```bash=
docker-compose up -d
```
This will launch five containers:
- `dgraph`'s database driver
- `dgraph`'s API server
- `dgraph`'s frontend, Ratel
- `wikidist`'s API
- `wikidist`'s front-end

In order to launch the crawler, see [here](#crawler).

It is an easy way to test the project. However, if you really want to go deeper into the project, you can still install the different component by hand.

## The long way

### [Dgraph](https://dgraph.io/)

Use `docker-compose` to deploy `dgraph`:

```bash=
docker-compose up -d dgraph-server dgraph-ratel dgraph-zero
```

This will launch the three containers used by `dgraph`:
- `dgraph`'s database driver
- `dgraph`'s API server
- `dgraph`'s frontend, Ratel

If for some reason you want to install it without using `docker-compose`, follow the instructions provided [here](https://docs.dgraph.io/deploy/).

### API

The API server uses port 8081. To launch it:

```bash=
cd cmd/api
go build
./api
```

### Front-end

#### Configuration

To change API url in frontend you can edit `frontend/.env`

By default the address is set to the VM hosted by CentraleSupélec, which can only be accessed from inside the CS network.

If you want to test locally, you can put the local API host adress and port (by default: `localhost:8081`)

``` js
VUE_APP_API_URL="http://localhost:8081"
```

#### Serving files

For development, you can serve the frontend with:

```
cd frontend
yarn install
yarn serve
```

#### Making a production build

For production, you should build static files using:

```
yarn build
```

### Crawler

The crawler only needs to be executed once in order to fill the database.


To launch the crawler:

```bash=
cd cmd/crawler
go build
./crawler <prefix> <start> <nb_worker>
```
- **prefix**: prefix of the Wikipedia instance to be crawled (e.g. 'en', 'fr', etc.)
- **start**: exact title of the page from which to start crawling (e.g. Alan Turing)
- **nb_worker**: number of concurrent workers to use while crawling (e.g. 5) - the optimal value depends on the setup (you should use monitoring tools to help you choose an appropriate one)

For example:

```bash=
./crawler fr "Alan Turing" 5
```

# Monitoring

The crawling performance is critical for this project considering the number of articles to fetch (for example, the French version of Wikipedia has upwards of 2 million articles). It was therefore essential to identify and fix bottlenecks to be able to crawl the whole graph in a reasonable amount of time.

We used Datadog to analyze metrics like queue length, articles fetched per second and cache hit ratio. Displaying those metrics in a meaningful and useful way helped us a lot in improving the crawling rate.

![monitoring](https://user-images.githubusercontent.com/8351433/74614060-e4eaa480-5114-11ea-9066-ce17ba4a953c.png)

# Testing

## Unit tests

The _fetcher_ part of package _crawler_ is unit tested.

All 100% of statements from `fetcher.go` are covered.

To run tests:

```sh
cd pkg/crawler
go test .
```

To show test coverage:

```sh
cd pkg/crawler
go test -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## CI

After each `git push`, the CI (Github Actions) will run the tests and report back on the appropriate PR whether tests are passing or not.

The actions executed by the CI are defined in [`.github/workflows/test.yml`](.github/workflows/test.yml).

# Trello / Scrum

In this project, we used Github Pull Requests instead of Trello.
Pull Requests are [available here](https://github.com/wikidistance/wikidist/pulls?utf8=%E2%9C%93&q=is%3Apr+).
