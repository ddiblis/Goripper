package main

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/gocolly/colly"
)

type info struct {
	name    string
	chNum   string
	pageNum string
}

func main() {
	//run("tate-no-yuusha-no-nariagari")
	run(os.Args[len(os.Args)-1])
}

func run(seriesName string) {

	var (
		chapwg sync.WaitGroup
		pagewg sync.WaitGroup
	)

	seriesURL := fmt.Sprintf("http://www.mangapanda.com/%s", seriesName)
	chanChapter := make(chan string)
	chanPage := make(chan string)
	chapterCollector := Chapters(seriesURL, chanChapter)

	go func() {
		defer close(chanPage)
		for chapURL := range chanChapter {
			chapwg.Add(1)
			go Pages(chapURL, chanPage, &chapwg)
		}
		chapwg.Wait()
	}()

	chapterCollector.Wait()
	// close(chanPage)

	for pageURL := range chanPage {
		pagewg.Add(1)
		go Images(pageURL, &pagewg)
	}

	pagewg.Wait()
}

func createInfo(pageURL string) (pageInfo *info) {

	structURL := strings.Split(pageURL, "/")
	if len(structURL) < 6 {
		structURL = append(structURL, "1")
	}

	pageInfo = &info{
		name:    structURL[3],
		chNum:   structURL[4],
		pageNum: structURL[5],
	}
	return
}

// Images recieves url to pages and searches for the link to the image
func Images(pageURL string, pagewg *sync.WaitGroup) {

	pageInfo := createInfo(pageURL)

	defer pagewg.Done()
	collector := colly.NewCollector(
		colly.Async(true),
	)

	collector.OnHTML("div#imgholder", func(e *colly.HTMLElement) {
		e.ForEach("img", func(_ int, o *colly.HTMLElement) {
			//fmt.Println(e.Request.AbsoluteURL(o.Attr("src")))
		})
	})

	collector.Visit(pageURL)
	collector.Wait()
	fmt.Printf("PageNumber: %v\n", pageInfo.pageNum)
}

// Pages : Parses html for img src link
func Pages(chapURL string, pg chan string, chapwg *sync.WaitGroup) {

	// defer fmt.Println("Closing Pages")
	defer chapwg.Done()
	collector := colly.NewCollector(
		colly.Async(true),
	)

	collector.OnHTML("select#pageMenu", func(e *colly.HTMLElement) {
		e.ForEach("option", func(_ int, o *colly.HTMLElement) {
			pg <- e.Request.AbsoluteURL(o.Attr("value"))
		})
	})

	collector.Visit(chapURL)
	collector.Wait()
}

// Chapters : parses the chapters and returns pages through a channel
func Chapters(seriesURL string, ch chan string) *colly.Collector {

	//seriesURL := "http://www.mangapanda.com/tate-no-yuusha-no-nariagari"

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
