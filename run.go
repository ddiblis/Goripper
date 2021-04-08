package main

import (
	"fmt"
	"sync"
)

// runOne is a generic go routine to be used with each of the channels in this package
func runOne(toClose chan string, toIter chan string, toRun func(string, chan string, *sync.WaitGroup), wg *sync.WaitGroup) {

	if toClose != nil {
		defer close(toClose)
	}

	for url := range toIter {
		wg.Add(1)
		go toRun(url, toClose, wg)
	}

	wg.Wait()
}

// run is the function controlling everything in this package
func run(seriesName string) {

	seriesURL := fmt.Sprintf("http://www.mangapanda.com/%s", seriesName)
	chanChapter := make(chan string)
	chanPage := make(chan string)
	chanImage := make(chan imageRef)
	chapterCollector := Chapters(seriesURL, chanChapter)

	var (
		chapwg  sync.WaitGroup
		pagewg  sync.WaitGroup
		imagewg sync.WaitGroup
	)

	workers := make(chan bool, 4)
	go func() {
		defer close(chanPage)
		for chapURL := range chanChapter {
			chapwg.Add(1)
			workers <- false
			go Pages(chapURL, chanPage, &chapwg, workers)
		}
		chapwg.Wait()
	}()

	go func() {
		defer close(chanImage)
		for pageURL := range chanPage {
			pagewg.Add(1)
			go Images(pageURL, chanImage, &pagewg)
		}
		pagewg.Wait()
	}()

	go func() {
		for imgURL := range chanImage {
			imagewg.Add(1)
			go DownloadImg(imgURL, &imagewg)
		}
	}()

	chapterCollector.Wait()
	imagewg.Wait()

	// go runOne(chanPage, chanChapter, Pages, &chapwg)
	// chapterCollector.Wait()
	// runOne(nil, chanPage, Images, &pagewg)

}
