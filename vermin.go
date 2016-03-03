package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/jackdanger/collectlinks"
)

// map for storing visited uri's to avoid visit loops.
var visited = make(map[string]bool)

// map for storing found dead links.
var deadLinks = make(map[string]error)

func main() {
	flag.Parse()

	args := flag.Args()

	fmt.Println("Start Page:", args)

	//check if an arg was passed along.
	if len(args) < 1 {
		fmt.Println("Specify a start page")
		os.Exit(1)
	}

	queue := make(chan string)

	//put arg into channel for queuing.
	go func() {
		queue <- args[0]
	}()
	for uri := range queue {
		queueLinks(uri, queue)
	}
	fmt.Println("visited: ", visited)
	//fmt.Println(deadLinks, "[Broken]")
}

// queue found links for processing and checking.
func queueLinks(uri string, queue chan string) {
	//fmt.Println("Fetching", uri)

	//store & tag uri as visited.
	visited[uri] = true
	httpTransport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	httpClient := http.Client{Transport: httpTransport}

	resp, err := httpClient.Get(uri)
	if err != nil {
		fmt.Println("resp Error: ", err)
		deadLinks["Possible dead link :"] = err
		fmt.Println(deadLinks)
		return
	}

	//fmt.Println("response body ", resp)
	defer resp.Body.Close()

	// collectlinks package helps in parsing a webpage & returning found
	// hyperlink href.
	links := collectlinks.All(resp.Body)
	//fmt.Println("links:", links)
	for _, link := range links {

		absolute := fixURL(link, uri)
		if uri != "" {
			//dont queue a uri twice.
			if !visited[absolute] {
				go func() { queue <- absolute }()
				//	fmt.Println("channel queue:", queue)
			}
		}
	}
}

//fix url
func fixURL(href, base string) string {
	uri, err := url.Parse(href)
	if err != nil {
		return ""
	}
	baseURL, err := url.Parse(base)
	if err != nil {
		return ""
	}
	uri = baseURL.ResolveReference(uri)
	//fmt.Println("fixed:", uri)
	return uri.String()
}
