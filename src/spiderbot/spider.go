package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// SitemapNode Node of the graph that represents the sitemap
type SitemapNode struct {
	root  string
	links map[*SitemapNode]bool
}

type pageURLs struct {
	root string
	urls map[string]bool
}

// Spider Spiderbot user to crawl a domain.
type Spider struct {
	rooturl  string
	client   *http.Client
	timeout  int
	maxDepth int
}

// MakeSpider Creates a new Spider.
func MakeSpider(rawurl string, timeout int, maxDepth int) (*Spider, error) {
	rooturl, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}
	// HTTP client with timeout (safe for concurrent use by multiple goroutines)
	client := http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}
	return &Spider{rooturl.String(), &client, timeout, maxDepth}, nil
}

// Helper function to get the content of an href attribute from an HTML token.
func getLink(t *html.Token) (string, error) {
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
func (s *Spider) parseURL(rawurl string, chPage chan<- pageURLs) {
	//log.Println("Crawling:", rawurl)
	page := pageURLs{rawurl, nil}
	// get the root URL raw path
	absolute, err := url.Parse(s.rooturl)
	if err != nil {
		message := fmt.Sprintf("[%s] Unable to parse the URL: %s", err, rawurl)
		log.Println(message)
		chPage <- page
		return
	}
	rootPath := absolute.Path
	resp, err := s.client.Get(rawurl)
	// check for requests errors
	if err != nil || resp.StatusCode != http.StatusOK {
		// TODO: verbose logging
		chPage <- page
		return
	}
	b := resp.Body
	defer b.Close()

	page.urls = make(map[string]bool)
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
			token := z.Token()
			link, err := getLink(&token)
			if err != nil {
				// TODO: verbose logging
				continue
			}
			// do not follow external links (or itself)
			if !strings.HasPrefix(link, "/") || link == "/" {
				if !strings.HasPrefix(link, s.rooturl) {
					continue
				} else {
					link = strings.Replace(link, s.rooturl, "", 1)
				}
			}
			// store the absolute URL path
			absolute.Path = path.Join(rootPath, link)
			page.urls[absolute.String()] = true
		}
	}
}

// Crawl Explore the site map without following external links.
func (s *Spider) Crawl() *SitemapNode {
	// list of urls to visit
	toVisit := []string{s.rooturl}
	// channel of visited urls with their content as urls list
	chPage := make(chan pageURLs)

	// sitemap graph
	rootNode := SitemapNode{s.rooturl, make(map[*SitemapNode]bool)}
	// map of visited nodes (link to the graph nodes and avoid visiting the same
	// nodes during the breadth first search)
	nodes := make(map[string]*SitemapNode)
	nodes[s.rooturl] = &rootNode

	depth := 0
	// modified version of breadth first search
	for len(toVisit) > 0 {
		// ensure a maximum "recursion" limit
		if depth > s.maxDepth && s.maxDepth > 0 {
			break
		}
		depth++

		// process in parallel all the current urls to visit
		length := len(toVisit)
		log.Printf("Going to visit %d URLs\n", length)
		i := 0
		for i < length {
			go s.parseURL(toVisit[i], chPage)
			i++
		}
		toVisit = toVisit[:0]

		// get results from all the goroutines
		log.Println("Waiting for results...")
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
			for url := range page.urls {
				// if the link is not stored in the graph not create a new node
				if _, exists := nodes[url]; !exists {
					nodes[url] = &SitemapNode{url, make(map[*SitemapNode]bool)}
					toVisit = append(toVisit, url)
				}
				// link url to the root node
				root.links[nodes[url]] = true
			}
			i++
		}
		log.Println("Visited", len(nodes))
		log.Println("To visit", len(toVisit))
	}

	return &rootNode
}

// Print Explore the sitemap graph with a breadth first search and prints it to
// the standard output.
func (sitemap *SitemapNode) Print() {
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
