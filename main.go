package main

import (
	"flag"
	"fmt"
	"golang.org/x/net/html"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
)

type Page struct {
	url    url.URL
	assets []string
}

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

	var wg sync.WaitGroup

	wg.Add(1)
	go Crawl(root, *depth, &wg)

	wg.Wait()
	fmt.Println("Wait over!")
}

func Crawl(u url.URL, depth int, wg *sync.WaitGroup) {
	defer wg.Done()
	if depth <= 0 {
		fmt.Println("Reached max depth")
		return
	}

	fmt.Printf("Crawling %s\n", u.String())

	resp, err := http.Get(u.String())
	if err != nil {
		panic(err)
	}

	page := Page{url: u, assets: make([]string, 0)}

	z := html.NewTokenizer(resp.Body)
	for {
		tt := z.Next()

		switch tt {
		case html.ErrorToken:
			fmt.Printf("Resources for %s\n", page.url.String())
			for _, v := range page.assets {
				fmt.Printf("* %s\n", v)
			}
			return
		case html.StartTagToken:
			t := z.Token()

			switch t.Data {
			case "link", "img", "script":
				for _, a := range t.Attr {
					if a.Key == "src" || a.Key == "href" {
						page.assets = append(page.assets, a.Val)
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
							wg.Add(1)
							go Crawl(*_u, depth-1, wg)
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
