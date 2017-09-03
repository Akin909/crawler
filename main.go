// Package main provides a simple web crawler
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"golang.org/x/net/html"
)

func crawl(url string, ch chan string, chFinished chan bool) {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal("Error: Failed to crawl \"" + url + "\"")
		log.Println(err)
	}

	defer func() {
		//Notify that we're done after this function finishes
		chFinished <- true
	}()

	b := resp.Body
	defer b.Close() //Close body when function returns

	//==========================================================
	// Parse HTML for anchor Tags
	//==========================================================

	z := html.NewTokenizer(b)

	for {
		tt := z.Next()

		switch {
		case tt == html.ErrorToken:
			//End of document
			return

		case tt == html.StartTagToken:
			t := z.Token()
			isAnchor := t.Data == "a"
			if !isAnchor {
				continue
			}
			OK, url := getHref(t)
			if !OK {
				log.Println("No go, carrying on")
				continue
			}

			//Make sure the url begins with http
			hasProto := strings.Index(url, "http") == 0
			if hasProto {
				ch <- url
			}
		}

		resp.Body.Close()
	}
}

func main() {
	foundUrls := make(map[string]bool)
	seedUrls := os.Args[1:]

	//Channels
	chUrls := make(chan string)
	chFinished := make(chan bool)

	//Start Concurret Crawl
	for _, url := range seedUrls {
		go crawl(url, chUrls, chFinished)
	}

	//Subscribe to both channels
	for c := 0; c < len(seedUrls); {
		select {
		case url := <-chUrls:
			foundUrls[url] = true
		case <-chFinished:
			c++
		}
	}
	//Print the results
	fmt.Println("\nFound", len(foundUrls), "unique urls: ")

	for url := range foundUrls {
		fmt.Println(" - " + url)
	}
	close(chUrls)
}

func getHref(t html.Token) (OK bool, href string) {
	for _, a := range t.Attr {
		if a.Key == "href" {
			href = a.Val
			OK = true
		}
	}
	return
}
