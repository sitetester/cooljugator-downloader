package main

import (
	"cooljugator_lt/entity"
	"cooljugator_lt/service"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const CooljugatorUrl = "https://cooljugator.com/"

func main() {
	resp, err := http.Get(CooljugatorUrl)
	if err != nil {
		panic(err)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Errorf("bad status: %s", resp.Status)
	}

	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	languageCodeNameMap := downloadLanguageCodes(*doc)
	linksByLanguageCode := make(chan entity.LinksByLanguageCode, len(languageCodeNameMap))
	for languageCode := range languageCodeNameMap {
		go parseEachLanguageCodeLinks(languageCode, linksByLanguageCode)
	}

	LinksByLanguageCodeMap := getLinksByLanguageCodeMap(linksByLanguageCode, len(languageCodeNameMap))

	var wgOuter sync.WaitGroup
	for languageCode, links := range LinksByLanguageCodeMap {
		wgOuter.Add(1)
		go downloadAllLanguageCodeLinks(languageCode, links, &wgOuter)
	}

	wgOuter.Wait()
	fmt.Println("All language codes parsed!")
}

func downloadAllLanguageCodeLinks(languageCode string, links []entity.Link, wgOuter *sync.WaitGroup) {
	defer wgOuter.Done()
	dir := "./languages" + string(filepath.Separator) + languageCode
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0777)
	}

	var downloader service.Downloader
	var wgInner sync.WaitGroup
	counter := 0

	for _, link := range links {
		wgInner.Add(1)
		go downloader.Download(dir, CooljugatorUrl+link.Href[1:], link.HrefText+".html", &wgInner)
		counter += 1

		if counter == 10 {
			counter = 0
			time.Sleep(time.Second * 5)
			fmt.Println("Sleeping....")
		}
	}

	wgInner.Wait()
	fmt.Println("All done!")
}

func getLinksByLanguageCodeMap(ch chan entity.LinksByLanguageCode, totalLanguageCodes int) map[string][]entity.Link {
	linksByLanguageCodeMap := make(map[string][]entity.Link)

	for {
		select {
		case linksByLanguageCode := <-ch:
			linksByLanguageCodeMap[linksByLanguageCode.LanguageCode] = linksByLanguageCode.Links

			if len(linksByLanguageCodeMap) == totalLanguageCodes {
				return linksByLanguageCodeMap
			}
		}
	}
}

func downloadLanguageCodes(doc goquery.Document) map[string]string {
	languageCodeNameMap := make(map[string]string)

	doc.Find("#main-language-selection a").Each(func(index int, item *goquery.Selection) {
		a := item
		href, _ := a.Attr("href")
		languageCode := href[1 : len(href)-1]
		name := a.Text()

		languageCodeNameMap[languageCode] = name
	})

	return languageCodeNameMap
}

func parseEachLanguageCodeLinks(languageCode string, linksByLanguageCodeChan chan entity.LinksByLanguageCode) {
	// some languages don't have `/all` page, only `/list`
	url := CooljugatorUrl + languageCode + "/list/all"
	fmt.Println("Parsing language code ...", url)

	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Errorf("bad status: %s", resp.Status)
	}

	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var links []entity.Link

	doc.Find("#conjugate ul li.item a").Each(func(index int, item *goquery.Selection) {
		linkTag := item
		href, _ := linkTag.Attr("href")
		linkText := linkTag.Text()

		var link entity.Link
		link.Href = href
		link.HrefText = linkText

		links = append(links, link)
	})

	// when links are fully parsed, put it into linksByLanguageCodeChan channel
	// LanguageCode will be used to identity which links are parsed
	linksByLanguageCodeChan <- entity.LinksByLanguageCode{LanguageCode: languageCode, Links: links}
}
