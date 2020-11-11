package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gocolly/colly"
	"github.com/schollz/progressbar/v3"
)

func main() {
	run(os.Args[len(os.Args)-1])
}

// createInfo extracts pageNUm, chapterNum, and name of series from the URL
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

// Chapters : parses the chapters and returns pages through a channel
func Chapters(seriesURL string, ch chan string) (collector *colly.Collector) {

	collector = colly.NewCollector(
		colly.Async(true),
	)

	// folderInfo := createInfo(chapURL)

	// os.MkdirAll()

	collector.OnHTML("table#listing", func(e *colly.HTMLElement) {
		defer close(ch)
		e.ForEach("a", func(_ int, o *colly.HTMLElement) {
			ch <- e.Request.AbsoluteURL(o.Attr("href"))
		})
	})

	collector.Visit(seriesURL)

	return
}

// Pages : Parses html for img src link
func Pages(chapURL string, pg chan string, chapwg *sync.WaitGroup) {

	defer chapwg.Done()

	info := createInfo(chapURL)
	path := info.name + "/" + info.chNum
	os.MkdirAll(path, os.FileMode(0755))

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

// Images recieves url to pages and searches for the link to the image
func Images(pageURL string, img chan imageRef, pagewg *sync.WaitGroup) {

	defer pagewg.Done()

	collector := colly.NewCollector(
		colly.Async(true),
	)

	collector.OnHTML("div#imgholder", func(e *colly.HTMLElement) {
		e.ForEach("img", func(_ int, o *colly.HTMLElement) {
			img <- imageRef{e.Request.AbsoluteURL(o.Attr("src")), pageURL}
		})
	})

	collector.Visit(pageURL)
	collector.Wait()
}

// DownloadImg downloads the image
func DownloadImg(imgURL imageRef, imagewg *sync.WaitGroup) {

	defer imagewg.Done()

	pageInfo := createInfo(imgURL.PageURL)
	path := pageInfo.name + "/" + pageInfo.chNum
	fileName := path + "/" + pageInfo.pageNum + ".jpg"

	resp, _ := http.Get(imgURL.ImgURL)

	defer resp.Body.Close()

	file, _ := os.Create(fileName)

	defer file.Close()

	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		fmt.Sprintf("Downloading %v", fileName),
	)

	io.Copy(io.MultiWriter(file, bar), resp.Body)
}
