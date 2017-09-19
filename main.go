package main

import (
	"container/list"
	"flag"
	"fmt"
	"golang.org/x/net/html"
	"net/http"
	"net/url"
	"os"
)

/* Prefer to make this const */
var (
	depth          *int  = flag.Int("depth", 2, "Crawl depth")
	followExternal *bool = flag.Bool("follow-external", false, "Follow external links")
	root           url.URL
)

func main() {
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Println("Args != 1")
		os.Exit(1)
	}

	url, err := url.ParseRequestURI(flag.Arg(0))
	if err != nil {
		fmt.Printf("Input was not a valid URL: %s\n", root.String())
		os.Exit(1)
	}

	root = *url

	go Crawl(root, *depth)

func BeginCrawl(u url.URL, maxDepth int) {
	Crawl(u, 0, maxDepth)
}

func Crawl(u url.URL, depth int) {
	if depth <= 0 {
		fmt.Println("Reached max depth")
		return
	}

	fmt.Printf("Crawling %s\n", u.String())

	resp, err := http.Get(u.String())
	if err != nil {
		panic(err)
	}

	z := html.NewTokenizer(resp.Body)

	resources := list.New()

	for {
		tt := z.Next()

		switch {
		case tt == html.ErrorToken:
			fmt.Printf("%s contained resources %v\n", u.Hostname(), resources)
			return
		case tt == html.StartTagToken:
			t := z.Token()

			if t.Data == "link" {
				for _, a := range t.Attr {
					if a.Key == "href" {
						resources.PushBack(a.Val)
					}
				}
			}

			if t.Data == "a" {
				for _, a := range t.Attr {
					if a.Key == "href" {
						_u, err := url.ParseRequestURI(a.Val)
						if err != nil {
							fmt.Println("Could not parse URL in href")
							return
						}

						if len(_u.Hostname()) == 0 {
							_u.Host = u.Hostname()
							_u.Scheme = u.Scheme
						}

						if shouldVisit(*_u) {
							fmt.Printf("Crawling %s\n", _u.String())
							go Crawl(*_u, depth-1)
						} else {
							fmt.Printf("Skipping %s\n", _u.Hostname())
						}
					}
				}
			}
		}
	}

}

func shouldVisit(u url.URL) bool {
	return ((u.Hostname() != root.Hostname()) && (*followExternal)) || (u.Hostname() == root.Hostname())
}

func usage() {
	fmt.Printf("Usage: %s <url>", os.Args[0])
}
