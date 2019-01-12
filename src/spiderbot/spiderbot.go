package main

import (
	"flag"
	"log"
)

func main() {
	// parse command line flags
	rooturl := flag.String("url", "https://monzo.com/", "URL to crawl")
	timeout := flag.Int("timeout", 5, "HTTP request timeout in seconds")
	maxDepth := flag.Int("max-depth", -1, "Max exploration depth, neg for unlimited")
	flag.Parse()

	// create the spiderbot
	log.Printf("Crawling %s\n", *rooturl)
	spider, err := MakeSpider(*rooturl, *timeout, *maxDepth)
	if err != nil {
		log.Fatal("Unable to create the spiderbot")
	}
	sitemap := spider.Crawl()
	log.Println("Crawling completed!")
	sitemap.Print()
}
