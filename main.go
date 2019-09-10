package main

import (
	"cooljugator_lt/service"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"sync"
	"time"
)

// TODO: Parse all links from this link
func main() {
	const URL = "https://cooljugator.com/lt/list/all"

	resp, err := http.Get(URL)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		fmt.Errorf("bad status: %s", resp.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	links := make(map[string]string)

	// use CSS selector found with the browser inspector
	// for each, use index and item
	doc.Find("#conjugate ul li.item a").Each(func(index int, item *goquery.Selection) {
		linkTag := item
		href, _ := linkTag.Attr("href")
		linkText := linkTag.Text()
		links[href] = linkText
	})

	var downloader service.Downloader
	var wg sync.WaitGroup

	counter := 0
	for href, linkText := range links {
		wg.Add(1)
		go downloader.Download("https://cooljugator.com"+href, linkText+".html", &wg)
		counter += 1

		if counter == 500 {
			counter = 0
			time.Sleep(time.Second * 10)
			fmt.Println("Sleeping....")
		}
	}

	wg.Wait()
	fmt.Println("All done!")
}
