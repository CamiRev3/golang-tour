package main

import (
	"fmt"
	"sync"
)

type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

type SafeCrawler struct {
	mu    sync.Mutex
	cache map[string]bool
}

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func (sc *SafeCrawler) Crawl(url string, depth int, fetcher Fetcher, ch chan string,
	cached chan string, branch chan int) {

	if depth <= 0 {
		return
	}

	sc.mu.Lock()

	exist := sc.cache[url]

	if exist {
		cached <- fmt.Sprint("url: ", url, " already cached")
		sc.mu.Unlock()
		return
	}

	sc.cache[url] = true
	sc.mu.Unlock()

	_, urls, err := fetcher.Fetch(url)

	branch <- len(urls)

	if err != nil {
		ch <- fmt.Sprint(err)
		return
	}

	for _, u := range urls {
		go sc.Crawl(u, depth-1, fetcher, ch, cached, branch)
	}
	ch <- fmt.Sprintf("found: %s\n", url)

}

func main() {
	sc := SafeCrawler{cache: make(map[string]bool)}

	ch := make(chan string)
	cached := make(chan string)
	branch := make(chan int)

	branchesPlusRoot := 1
	

	go sc.Crawl("https://golang.org/", 4, fetcher, ch, cached, branch)

	for founded := 0; founded < branchesPlusRoot; {

		select {
		case newUrl := <-ch:
			fmt.Println("new url ->", newUrl)
			founded++
		case <-cached:
			founded++
		case branches := <-branch:
			branchesPlusRoot += branches
		}

	}
}

// fakeFetcher is Fetcher that returns canned results.
type fakeFetcher map[string]*fakeResult

type fakeResult struct {
	body string
	urls []string
}

func (f fakeFetcher) Fetch(url string) (string, []string, error) {
	if res, ok := f[url]; ok {
		return res.body, res.urls, nil
	}
	return "", nil, fmt.Errorf("not found: %s", url)
}

// fetcher is a populated fakeFetcher.
var fetcher = fakeFetcher{
	"https://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"https://golang.org/pkg/",
			"https://golang.org/cmd/",
		},
	},
	"https://golang.org/pkg/": &fakeResult{
		"Packages",
		[]string{
			"https://golang.org/",
			"https://golang.org/cmd/",
			"https://golang.org/pkg/fmt/",
			"https://golang.org/pkg/os/",
		},
	},
	"https://golang.org/pkg/fmt/": &fakeResult{
		"Package fmt",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
	"https://golang.org/pkg/os/": &fakeResult{
		"Package os",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
}
