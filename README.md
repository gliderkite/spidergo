# Web Crawler

The goal of this project is to provide a simple implementation of a web crawler
(spiderbot), which will be:

- Limited to a single domain (without following external links).

should be limited to one domain - so when you start with https://monzo.com/, it 
would crawl all pages within monzo.com, but not follow external links, for 
example to the Facebook and Twitter accounts. Given a URL, it should print a 
simple site map, showing the links between pages.

write it as you would a production piece of code. Bonus points for tests and 
making it as fast as possible!

we care less about a fancy UI or rendering the resulting sitemap nicely and more 
about how your program is structured, the trade-offs you've made, what behaviour
the program exhibits etc..

*Disclaimer: this is the first program I write in Go.*


# Design and Implementation

- The program will accept the domain to crawl as the first command line argument
which will be the starting (url) *seed*.
- For each url, we get through an HTTP request the content of the body of the
webpage, which will be parsed and all its link will be extracted. All these new
links need to be added to the list of links to visit (the crawl frontier), but
it is important to check if a link has already been visited or not: for this
purpose the algorithm will be bread first search (this strategy is supported
also by https://oak.cs.ucla.edu/~cho/papers/cho-thesis.pdf: the crawler wants to
download pages with high Pagerank early during the crawling process, then the
partial Pagerank strategy is the better, followed by breadth-first and
backlink-count.)
- The same procedure will be applied recursively for each non-external link
found in each webpage, therefore a link parser that will check if the link is
external or not is required. 
https://siongui.github.io/2018/01/31/go-get-domain-name-from-url/
https://play.golang.org/p/AmtdM9lWK-x
- Beside the queue and set data structures needed for the links exploration, a
separate data structure is required in order to build the site map. This data
structure could simply be a list of nodes.


# Usage



