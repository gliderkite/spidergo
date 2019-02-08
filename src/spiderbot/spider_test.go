package main

import (
	"testing"
)

// Tests that a spiderbot can be constructed with valid arguments.
func TestShouldConstructSpider(t *testing.T) {
	rooturl := "https://google.com/"
	timeout := 5
	maxDepth := 10
	maxUrls := 100
	spider, err := MakeSpider(rooturl, timeout, maxDepth, maxUrls)

	if err != nil {
		t.Error("Unable to construct a new spiderbot")
	} else if spider.rooturl != rooturl || spider.timeout != timeout || spider.maxDepth != maxDepth {
		t.Error("Spiderbot constructed with invalid arguments")
	}
}

// Tests that the spiderbot can correctly parse a given URL.
func TestShouldParseURL(t *testing.T) {
	rooturl := "https://google.com/"
	spider, err := MakeSpider(rooturl, 1, 1, 50)
	if err != nil {
		t.Error("Unable to construct a new spiderbot")
	}

	chPage := make(chan pageURLs)
	rawurl := rooturl + "services"
	go spider.parseURL(rawurl, chPage)
	page := <-chPage

	if page.root != rawurl || len(page.urls) < 1 {
		t.Error("Unable to parse URL")
	}
}

// Tests that the spiderbot can correctly crawl a given URL.
func TestShouldCrawl(t *testing.T) {
	rooturl := "https://google.com/"
	spider, err := MakeSpider(rooturl, 1, 1, 50)
	if err != nil {
		t.Error("Unable to construct a new spiderbot")
	}

	sitemap := spider.Crawl()

	if sitemap.root != rooturl || len(sitemap.links) < 1 {
		t.Error("Unable to crawl URL")
	}
}
