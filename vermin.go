package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jackdanger/collectlinks"
)

// visited is a map for storing visited uri's to avoid visit loops.
var visited = make(map[string]bool)

//  deadLinks is a slice for storing found dead links.
var deadLinks = []error{}

/*var httpTransport = &http.Transport{
	TLSClientConfig: &tls.Config{
		InsecureSkipVerify: true,
	},
}*/

var httpClient http.Client

//main function
func main() {

	var link = flag.String("url", "", "Url to crawl for dead links")
	var boolflag = flag.Bool("hostonly", true, "flag to crawl none base host. Defaults to true")

	flag.Parse()

	queue := make(chan string)

	fmt.Println("flag value:", *link)

	// Go routine puts url flag into channel for queuing.
	go func() {
		queue <- *link
	}()

	hostURL, err := url.Parse(*link)
	if err != nil {
		//fmt.Println("Invalid Url: ", err)
		return
	}

	for {
		select {
		case uri, ok := <-queue:
			if !ok {
				return
			}

			queueLinks(hostURL.Host, uri, queue, *boolflag)

		case <-time.After(1 * time.Minute):
			fmt.Println("Ending crawler. Goodbye!")
			fmt.Println("--------------------DEAD LINKS------------------------------")

			for _, f := range deadLinks {
				fmt.Println(f)
			}

			fmt.Println("------------------------------------------------------------")

			return
		}
	}

	// for uri := range queue {
	// 	queueLinks(hostURL.Host, uri, queue, *boolflag)
	// }

}

// queueLinks is used for making http calls to the queued links in queue channel

func queueLinks(host, uri string, queue chan string, boolflag bool) {
	if visited[uri] {
		return
	}

	crawledURL, _ := url.Parse(uri)

	if boolflag {
		if !strings.Contains(crawledURL.Host, host) {
			// fmt.Println("Current url host name doesn't match base url host name.")
			return
		}
	}

	fmt.Printf("Fetching: %s\n", uri)

	// store & tag uri as visited.
	visited[uri] = true

	resp, err := httpClient.Get(uri)
	if err != nil {
		fmt.Printf(`Host: %s
URI: %s
Error: %s

`, host, uri, err.Error())

		// fmt.Println("resp Error: ", err)
		//deadLinks["Possible dead link :"] = err
		deadLinks = append(deadLinks, err)
		return
	}

	fmt.Println("")

	//fmt.Println("response body ", resp)
	defer resp.Body.Close()

	// collectlinks package helps in parsing a webpage & returning found
	// hyperlink href.
	links := collectlinks.All(resp.Body)

	for _, link := range links {

		absolute, err := fixURL(link, uri)
		if err != nil {
			fmt.Println("link error: ", err)
			continue
		}

		// dont queue a uri twice.
		go func() {
			queue <- absolute
		}()
	}
}

// fix url
func fixURL(href, base string) (string, error) {
	uri, err := url.Parse(href)
	if err != nil {
		return "", errors.New("link error")
	}

	baseURL, err := url.Parse(base)
	if err != nil {
		return "", errors.New("Link error")
	}

	uri = baseURL.ResolveReference(uri)
	//fmt.Println("uri without string: ", uri)

	/*if !boolflag {
		if uri.Host != baseURL.Host {
			return "", errors.New("External Link")
		}
	}*/

	//fmt.Println("resolvedURL:", uri.String())
	return uri.String(), err
}
