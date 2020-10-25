package main

import (
	"fmt"

	"github.com/gocolly/colly"
)

func Main() {
	Pages("https://www.mangareader.net/tate-no-yuusha-no-nariagari/1")

	// channel := make(chan string)
	// Chapters(channel)
	// insert for loop here using channel
	// increment wg then go Pages(page, wg)
	// wg.Wait()
}

func Pages(pageURL string) {

	c := colly.NewCollector(
		colly.Async(true),
	)

	c.OnHTML("table.d48", func(e *colly.HTMLElement) {
		e.ForEach("a", func(_ int, o *colly.HTMLElement) {
			fmt.Println(e.Request.AbsoluteURL(o.Attr("href")))
		})
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	c.Visit()

	c.Wait()

}

// Chapters parses the chapters and returns pages through a channel
func Chapters() {

	baseURL := "https://www.mangareader.net/tate-no-yuusha-no-nariagari"

	c := colly.NewCollector(
		colly.Async(true),
	)

	c.OnHTML("table.d48", func(e *colly.HTMLElement) {
		e.ForEach("a", func(_ int, o *colly.HTMLElement) {
			fmt.Println(e.Request.AbsoluteURL(o.Attr("href")))
		})
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	c.Visit(baseURL)

	c.Wait()
}
