/*
Do check out below elegant solution
http://grandiloquentmusings.blogspot.sg/2013/12/my-solution-to-go-tutorial-web-crawler.html
Tidied up my own solution (the locking of visited url part) after reading above solution.
*/
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

type visitedMap struct {
	visited map[string]bool
	lock    sync.Mutex
}

func (m *visitedMap) testAndSetVisit(url string) bool {
	defer func() {
		m.visited[url] = true
		m.lock.Unlock()
	}()
	m.lock.Lock()
	return m.visited[url]
}

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func Crawl(url string, depth int, fetcher Fetcher, v *visitedMap, done chan bool) {
	defer func() {
		done <- true
	}()

	if depth <= 0 || v.testAndSetVisit(url) {
		return
	}

	body, urls, err := fetcher.Fetch(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("found: %s %q\n", url, body)
	childDone := make(chan bool)
	for _, u := range urls {
		go Crawl(u, depth-1, fetcher, v, childDone)
	}

	for i := 0; i < len(urls); i++ {
		<-childDone
	}
	return
}

func main() {
	done := make(chan bool)
	visitedUrl := &visitedMap{visited: make(map[string]bool)}
	go Crawl("http://golang.org/", 4, fetcher, visitedUrl, done)
	<-done
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
	"http://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"http://golang.org/pkg/",
			"http://golang.org/cmd/",
		},
	},
	"http://golang.org/pkg/": &fakeResult{
		"Packages",
		[]string{
			"http://golang.org/",
			"http://golang.org/cmd/",
			"http://golang.org/pkg/fmt/",
			"http://golang.org/pkg/os/",
		},
	},
	"http://golang.org/pkg/fmt/": &fakeResult{
		"Package fmt",
		[]string{
			"http://golang.org/",
			"http://golang.org/pkg/",
		},
	},
	"http://golang.org/pkg/os/": &fakeResult{
		"Package os",
		[]string{
			"http://golang.org/",
			"http://golang.org/pkg/",
		},
	},
}
