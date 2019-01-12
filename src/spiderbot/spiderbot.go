package main

import (
	"flag"
	"log"
	"net/url"
)

func main() {
	// parse command line flags
	rooturl := flag.String("url", "https://monzo.com/", "URL to crawl")
	timeout := flag.Int("timeout", 5, "HTTP request timeout in seconds")
	maxDepth := flag.Int("max-depth", -1, "Max exploration depth, neg for unlimited")
	flag.Parse()

	rawurl, err := url.Parse(*rooturl)
	if err != nil {
		log.Fatal("The URL provided is not valid!")
	}

	// create the spiderbot
	log.Printf("Crawling %s\n", rawurl.String())
	spider := MakeSpider(rawurl.String(), *timeout, *maxDepth)
	sitemap := spider.Crawl()
	log.Println("Crawling completed!")
	sitemap.Print()
}
