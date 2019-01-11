package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"

	"golang.org/x/net/html"
)

type SitemapNode struct {
	root  string
	links map[*SitemapNode]bool
}

type PageURLs struct {
	root string
	urls []string
}

// Helper function to get the content of an href attribute from an HTML token.
func getLink(t html.Token) (string, error) {
	if t.Data != "a" {
		return "", errors.New("The given token is not a valid anchor")
	}
	for _, a := range t.Attr {
		if a.Key == "href" {
			return a.Val, nil
		}
	}
	return "", errors.New("Unable to find any href attribute")
}

// Given the raw URL sends all the found URLs in the webpage through the given
// channel.
func parseUrl(rooturl string, rawurl string, chPage chan PageURLs) {
	//fmt.Println("Crawling:", rawurl)
	page := PageURLs{rawurl, nil}
	// get the root URL raw path
	absolute, err := url.Parse(rooturl)
	if err != nil {
		message := fmt.Sprintf("[%s] Unable to parse the URL: %s", err, rawurl)
		log.Println(message)
		chPage <- page
		return
	}
	rootPath := absolute.Path
	resp, err := http.Get(rawurl)
	// check for requests errors
	if err != nil || resp.StatusCode != http.StatusOK {
		// TODO: verbose logging
		chPage <- page
		return
	}
	b := resp.Body
	defer b.Close()

	// list of URLs found in the webpage
	//var urls []string
	// parse the response body
	z := html.NewTokenizer(b)
	for {
		tt := z.Next()
		switch {
		case tt == html.ErrorToken:
			// EOF
			chPage <- page
			return
		case tt == html.StartTagToken:
			link, err := getLink(z.Token())
			if err != nil {
				// TODO: verbose logging
				continue
			}
			// do not follow external links (or itself)
			if !strings.HasPrefix(link, "/") || link == "/" {
				continue
			}
			// store the absolute URL path
			absolute.Path = path.Join(rootPath, link)
			page.urls = append(page.urls, absolute.String())
		}
	}
}

// Explore the site map without following external links.
func crawl(rooturl string, maxDepth int) SitemapNode {
	// list of urls to visit
	toVisit := []string{rooturl}
	// channel of visited urls with their content as urls list
	chPage := make(chan PageURLs)

	// sitemap graph
	rootNode := SitemapNode{rooturl, make(map[*SitemapNode]bool)}
	// map of visited nodes (link to the graph nodes and avoid visiting the same
	// nodes during the breadth first search)
	nodes := make(map[string]*SitemapNode)
	nodes[rooturl] = &rootNode

	depth := 0
	// breadth first search
	for len(toVisit) > 0 {
		if depth > maxDepth && maxDepth > 0 {
			break
		}
		depth++

		// process in parallel all the current urls to visit
		length := len(toVisit)
		i := 0
		for i < length {
			go parseUrl(rooturl, toVisit[i], chPage)
			i++
		}
		toVisit = nil

		// get results from all the goroutines
		i = 0
		for i < length {
			page := <-chPage
			// check if the page is valid
			if len(page.urls) == 0 {
				i++
				continue
			}
			// find root node
			root := nodes[page.root]
			// iterate over each URL found in the page
			for j := range page.urls {
				url := &page.urls[j]
				// if the link is not stored in the graph not create a new node
				if _, exists := nodes[*url]; !exists {
					nodes[*url] = &SitemapNode{*url, make(map[*SitemapNode]bool)}
					toVisit = append(toVisit, *url)
				}
				// link url to the root node
				root.links[nodes[*url]] = true
			}
			i++
		}

		fmt.Println("visited", len(nodes))
	}

	return rootNode
}

// Explore the sitemap graph with a breadth first search and prints it to the
// standard output.
func print(sitemap *SitemapNode) {
	queue := []*SitemapNode{sitemap}
	visited := make(map[string]bool)
	// bfs
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		fmt.Println(node.root)
		for link := range node.links {
			fmt.Printf("\t%s\n", link.root)
			if _, exists := visited[link.root]; !exists {
				queue = append(queue, link)
				visited[link.root] = true
			}
		}
	}
}

func main() {
	rooturl := "https://monzo.com/"
	maxDepth := -1
	sitemap := crawl(rooturl, maxDepth)
	print(&sitemap)
}
