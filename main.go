package main

import (
	"flag"
	"fmt"
	"golang.org/x/net/html"
	"net/http"
	"net/url"
	"strings"
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
		fmt.Println("Specify one URL to crawl")
		os.Exit(1)
	}

	url, err := url.ParseRequestURI(flag.Arg(0))
	if err != nil {
		fmt.Printf("Input was not a valid URL: %s\n", flag.Arg(0))
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

	var resources []string

	for {
		tt := z.Next()

		switch tt {
		case html.ErrorToken:
			fmt.Printf("Resources for %s\n", u.String())
			for _, v := range resources {
				fmt.Println(v)
			}
			return
		case html.StartTagToken:
			t := z.Token()

			switch t.Data {
			case "link":
				for _, a := range t.Attr {
					if a.Key == "href" {
						resources = append(resources, a.Val)
					}
				}

			case "img", "script":
				for _, a := range t.Attr {
					if a.Key == "src" {
						resources = append(resources, a.Val)
					}
				}

			case "a":
				for _, a := range t.Attr {
					if a.Key == "href" {
						_u, err := url.ParseRequestURI(a.Val)
						if err != nil {
							fmt.Println("Could not parse URL in href")
							return
						}

						resolveRelativeUrl(u, _u)

						if shouldVisit(_u) {
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

func resolveRelativeUrl(parent url.URL, child *url.URL) {
	if len(child.Hostname()) == 0 {
		child.Host = parent.Host
		child.Scheme = parent.Scheme
	}

	if !strings.HasPrefix(child.Path, "/") {
		child.Path = parent.Path + child.Path
	}
}

func shouldVisit(u *url.URL) bool {
	return ((u.Hostname() != root.Hostname()) && (*followExternal)) || (u.Hostname() == root.Hostname())
}

func usage() {
	fmt.Printf("Usage: %s <url>", os.Args[0])
}
