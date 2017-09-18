package main

import (
	"flag"
	"fmt"
	"golang.org/x/net/html"
	"net/http"
	"net/url"
	"os"
)

/* Prefer to make this const */
var (
	maxDepth       *int  = flag.Int("max-depth", 2, "Max crawl depth")
	followExternal *bool = flag.Bool("follow-external", false, "Follow external links")
	root           string
)

func main() {
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Println("Args != 1")
		os.Exit(1)
	}
	root := flag.Arg(0)

	url, err := url.ParseRequestURI(root)
	if err != nil {
		fmt.Printf("Input was not a valid URL: %s\n", root)
		os.Exit(1)
	}

	BeginCrawl(*url, *maxDepth)
}

func BeginCrawl(u url.URL, maxDepth int) {
	Crawl(u, 0, maxDepth)
}

func Crawl(u url.URL, depth int, maxDepth int) {
	if maxDepth == depth {
		fmt.Println("Reached max depth")
		return
	}

	fmt.Printf("Crawling %s\n", u.String())

	resp, err := http.Get(u.String())
	if err != nil {
		panic(err)
	}

	z := html.NewTokenizer(resp.Body)

	depth++
	for {
		tt := z.Next()

		switch {
		case tt == html.ErrorToken:
			return
		case tt == html.StartTagToken:
			t := z.Token()

			if t.Data == "a" {
				for _, a := range t.Attr {
					if a.Key == "href" {
						_u, err := url.ParseRequestURI(a.Val)
						if err != nil {
							fmt.Println("Not a real URL")
							return
						}

						fmt.Printf("[%s]\t%s\t %s\n", _u.String(), _u.Host, u.Host)
						if shouldVisit(*_u) {
							Crawl(*_u, depth, maxDepth)
						}
					}
				}
			}
		}
	}
}

func shouldVisit(u url.URL) bool {
	return (len(u.Hostname()) > 0) && (((u.Hostname() != root) && (*followExternal)) || (u.Hostname() == root))
}

func usage() {
	fmt.Printf("Usage: %s <url>", os.Args[0])
}
