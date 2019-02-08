# SpiderGo

The goal of this project is to provide a simple implementation of a web crawler
(spiderbot), which will be limited to a single domain (without following
external links).
Given a URL, it should print a simple site map, showing the links between pages.


## Overview

The project has a file used as the entry point for the application
`spider.go` that uses the functionalities provided by the data structures
defined in `spiderbot.go` (where the implementation of the web crawler resides).
Moreover, a file for unit testing, `spider_test.go`, is also provided within the
same folder.

The main entry point has the scope of parsing the command line flags,
constructing a new instance of the web crawler with the configuration provided,
and finally crawling the given URL.

The command line arguments that can be provided are:
- URL (optional): the URL (*seed*) to crawl (default: https://google.com/)
- Timeout (optional): the timeout for the HTTP requests (default: 5s)
- Max depth (optional): maximum depth used for the exploration of the sitemap.
    (default: unlimited)

Disclaimer: This web crawler does not yet comply to the [robots exclusion
standard](https://en.wikipedia.org/wiki/Robots_exclusion_standard).


## Requirements

In order to run the application, you need to install [The Go Programming
Language](https://golang.org/), and optionally
[GNU Make](https://www.gnu.org/software/make/), which provides a faster way to
build and run the application (as well as the unit tests).

You will also need to get the package `golang.org/x/net/html`, used as a
dependency of the `spidergo` project; you can get it with:

```bash
go get golang.org/x/net/html
```

## Usage

In order to build the application, you just need to run the command `make`, which
will build the executable `./spiderbot`. Running `make clean` will delete the
executable generated if it exists.

Once the executable has been generated, it is possible to run the web crawler
(with the default arguments) by simply running: `./spiderbot`. The arguments
can be specified as, for example:

```bash
./spiderbot -url=http://spotify.com -timeout=2 -max-depth=3
```

In order to run the unit test, simply run `make test`.

To have a complete list of the commands used check out the `Makefile`.

## Design and Implementation

The main data structure is defined by the class `Spider`, which can be easily
constructed by passing the configuration arguments (described above) to the
function:

```go
func MakeSpider(rawurl string, timeout int, maxDepth int) (*Spider, error)
```

The function `MakeSpider` returns a newly constructed instance of the `Spider`
type and an error indicating success or failure.

The type `Spider` exposes an API that can be used to crawl the specified URL,
which returns the root of a graph data structure that represents the sitemap:

```go
func (s *Spider) Crawl() *SitemapNode
```

The `SitemapNode` can be explored as a graph by following its links (a set of
pointers to other nodes).

```go
type SitemapNode struct {
	root  string
	links map[*SitemapNode]bool
}
```

In order to crawl the given URL, I based the exploration on a breadth-first
search algorithm variant. The choice of this algorithm is supported by the fact
that [if the crawler wants to download pages with high Pagerank early during the
crawling process, then the partial Pagerank strategy is the better one, followed by
breadth-first and backlink-count][1], and also because [breadth-first crawl captures
pages with high Pagerank early in the crawl][2].

In pseudo-code:

```python
frontier.enqueue(seed)
while len(frontier) > 0:
    length = len(frontier)
    for url in frontier:
        go parse(url, channel)
    frontier.clear()
    for i in range(length):
        page <- channel
        for url in page.urls:
            sitemap.link(page.root, url)
            if url not in visited:
                frontier.enqueue(url)
```

At each iteration, the algorithm launches a new goroutine for each new URL stored
in the frontier that has yet to be explored (in the first step, the only URL will be
the seed). The objective of the goroutine is to parse the given URL, by
retrieving the body of the webpage through a GET request. Once the content of
the page has been downloaded, it will be parsed by going through each HTML node and
filtering the `href` attribute by excluding external links. All the URLs are stored
in a set to avoid duplication (an URL will likely contain multiple links to the
same webpages).
The communication between the parsing goroutine and the crawling algorithm is
done through a channel that sends and receives objects of types:

```go
type pageURLs struct {
	root string
	urls map[string]bool
}
```

The `root` field is necessary in order to be able to quickly retrieve the
parent URL once the data is received.
Once all the goroutine have been scheduled, the iteration of the loop waits on
the channel to return all the URLs found. For each URL, it checks if that URL has
already been visited and, if not, the URL is added to the frontier of the URLs that will
be visited at the next iteration of the loop.
The exploration will stop either when there are no more URLs to visit, or when
the max exploration depth has been reached.

While the URLs are retrieved from the channel, the graph representing the
sitemap (defined above by the type `SitemapNode`) is also constructed, and the
crawling algorithm is optimized by needing only to use the sitemap graph to check
whether the URLs have been already visited or not.

In order to print the sitemap, the `SitemapNode` type exposes an API:

```go
func (sitemap *SitemapNode) Print()
```

This method implements a breadth-first search exploration of the graph, which
will print each visited URL to the standard output, followed by each URL found
in its webpage.

[1]: https://en.wikipedia.org/wiki/Web_crawler#cite_note-11
[2]: https://en.wikipedia.org/wiki/Web_crawler#cite_note-12
