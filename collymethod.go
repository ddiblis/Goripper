package main

import (
	"fmt"

	"sync"

	"github.com/gocolly/colly"
)

func main() {

	chanchapter := make(chan string)

	chanpage := make(chan string)

	chapterCollector := Chapters(chanchapter)

	pageCollector := Pages(chanpage)

	var (
		chapwg sync.WaitGroup
		pagewg sync.WaitGroup
	)

	for chapurl := range chanchapter {
		chapwg.Add(1)
		go Pages(chapurl, &chapwg)
	}

	chapterCollector.Wait()
	chapwg.Wait()

	for pageurl := range chanpage {
		pagewg.Add(1)
		go Images(pageurl, &pagewg)
	}

	pageCollector.Wait()
	pagewg.Wait()
}

// ImagesChannel : Fed img src from function Pages then creates a channel for page image urls
// func ImagesChannel() {
//
// }
//
// func Download_Image() {
//
// }

// Images recieves url to pages and searches for the link to the image
func Images(pageURL string, pagewg *sync.WaitGroup) {

	defer pagewg.Done()
	collector := colly.NewCollector(
		colly.Async(true),
	)

	collector.OnHTML("div#imgholder", func(e *colly.HTMLElement) {
		e.ForEach("img", func(_ int, o *colly.HTMLElement) {
			fmt.Println(e.Request.AbsoluteURL(o.Attr("src")))
		})
	})

	collector.Visit(pageURL)

	collector.Wait()

}

// Pages : Parses html for img src link
func Pages(pg chan string, chapwg *sync.WaitGroup) *colly.Collector {

	defer chapwg.Done()
	collector := colly.NewCollector(
		colly.Async(true),
	)

	collector.OnHTML("select#pageMenu", func(e *colly.HTMLElement) {
		e.ForEach("option", func(_ int, o *colly.HTMLElement) {
			pg <- e.Request.AbsoluteURL(o.Attr("value"))
		})
	})

	collector.Visit(chapterURL)

	collector.Wait()

	return collector

}

// Chapters : parses the chapters and returns pages through a channel
func Chapters(ch chan string) *colly.Collector {

	seriesURL := "http://www.mangapanda.com/tate-no-yuusha-no-nariagari"

	collector := colly.NewCollector(
		colly.Async(true),
	)

	collector.OnHTML("table#listing", func(e *colly.HTMLElement) {
		defer close(ch)
		e.ForEach("a", func(_ int, o *colly.HTMLElement) {
			ch <- e.Request.AbsoluteURL(o.Attr("href"))
		})
	})

	collector.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	collector.Visit(seriesURL)

	return collector
}
