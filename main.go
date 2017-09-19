package main

import (
	"flag"
	"fmt"
	"github.com/patrickmn/go-cache"
	"golang.org/x/net/html"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
)

type Page struct {
	url    url.URL
	links  []url.URL
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

	pageCache := cache.New(cache.NoExpiration, cache.NoExpiration)

	var wg sync.WaitGroup

	wg.Add(1)
	go Crawl(root, *depth, &wg, pageCache)

	wg.Wait()
	fmt.Println("Wait over!")

	displaySiteMap(*pageCache)
}

func Crawl(u url.URL, depth int, wg *sync.WaitGroup, pageCache *cache.Cache) {
	defer wg.Done()

	if depth <= 0 {
		fmt.Println("Reached max depth")
		return
	}

	resp, err := http.Get(u.String())
	if err != nil {
		panic(err)
	}

	page := Page{url: u, links: make([]url.URL, 0), assets: make([]string, 0)}

	z := html.NewTokenizer(resp.Body)
	for {
		tt := z.Next()

		switch tt {
		case html.ErrorToken:
			pageCache.Add(page.url.String(), page, cache.NoExpiration)
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
						page.links = append(page.links, *_u)

						if shouldVisit(_u, pageCache) {
							fmt.Printf("Visiting %s\n", _u.String())
							wg.Add(1)
							go Crawl(*_u, depth-1, wg, pageCache)
						}
					}
				}
			}
		}
	}
}

func resolveRelativeUrl(parent url.URL, child *url.URL) {
	/* Create a complete URL we can crawl */
	if len(child.Hostname()) == 0 {
		child.Host = parent.Host
		child.Scheme = parent.Scheme
	}

	/* Relative child paths should include parent and child paths */
	if !strings.HasPrefix(child.Path, "/") {
		child.Path = parent.Path + child.Path
	}
}

func shouldVisit(u *url.URL, pageCache *cache.Cache) bool {
	internal := u.Hostname() == root.Hostname()
	_, cacheHit := pageCache.Get(u.String())

	return (*followExternal || internal) && !cacheHit
}

func displaySiteMap(c cache.Cache) {
	for _, v := range c.Items() {
		page := v.Object.(Page)
		fmt.Printf("[%s]\n", page.url.String())

		fmt.Printf("Links:\n")
		for _, i := range page.links {
			fmt.Printf("\t%s\n", i.String())
		}

		fmt.Printf("Assets:\n")
		for _, i := range page.assets {
			fmt.Printf("\t%s\n", i)
		}
		fmt.Println("")
	}
}

func usage() {
	fmt.Printf("Usage: %s <url>", os.Args[0])
}
