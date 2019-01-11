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
	links []*SitemapNode
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

// Given the root URL sends all the found URLs in the webpage through the given
// channel.
func parseUrl(rooturl string, chPage chan PageURLs) {
	//fmt.Println("Crawling:", rooturl)
	page := PageURLs{rooturl, nil}
	// get the root URL raw path
	absolute, err := url.Parse(rooturl)
	if err != nil {
		message := fmt.Sprintf("[%s] Unable to parse the URL: %s", err, rooturl)
		log.Println(message)
		chPage <- page
		return
	}
	rootPath := absolute.Path
	resp, err := http.Get(rooturl)
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
func crawl(rooturl string) {
	// list of urls to visit and set of the onces already visited to avoid
	// duplicates
	toVisit := []string{rooturl}
	visited := make(map[string]bool)
	visited[rooturl] = true

	// channel of visited urls with their content as urls list
	chPage := make(chan PageURLs)

	maxDepth := 0
	// bread first search
	for len(toVisit) > 0 {
		if maxDepth > 2 {
			//break
		}
		maxDepth++

		// process in parallel all the current urls to visit
		length := len(toVisit)
		i := 0
		for i < length {
			go parseUrl(toVisit[i], chPage)
			i++
		}
		toVisit = nil

		// get results from all the goroutines
		i = 0
		for i < length {
			page := <-chPage
			// iterate over each URL found in the page
			for _, u := range page.urls {
				// check if already visited
				if _, exists := visited[u]; !exists {
					toVisit = append(toVisit, u)
					visited[u] = true
				}
			}
			i++
		}

		fmt.Println("visited", len(visited))
	}
}

func main() {
	rooturl := "http://monzo.com/"
	crawl(rooturl)
}
