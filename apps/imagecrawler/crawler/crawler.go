package crawler

import (
	"awesomeProject1/apps/imagecrawler/extracter"
	"awesomeProject1/apps/imagecrawler/queue"
	"awesomeProject1/db"
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type Crawler struct {
	urlQueue         *queue.URLQueue
	db               db.DB
	timeout          uint
	imagesFolderName string
	visited          *sync.Map
	sem              chan struct{}
	wg               *sync.WaitGroup
}

func NewCrawler(urlQueue *queue.URLQueue, db db.DB, timeout uint, imagesFolderName string, visited *sync.Map, sem chan struct{}, wg *sync.WaitGroup) *Crawler {
	return &Crawler{
		urlQueue:         urlQueue,
		db:               db,
		timeout:          timeout,
		imagesFolderName: imagesFolderName,
		visited:          visited,
		sem:              sem,
		wg:               wg,
	}
}

func (c *Crawler) Crawl() {
	defer c.wg.Done()

	for {
		URL, ok := c.urlQueue.Dequeue()
		if !ok {
			return
		}

		// Try to acquire a token from the semaphore
		c.sem <- struct{}{}

		// Create a new context for the current page with the specified timeout
		_, alreadyVisited := c.visited.LoadOrStore(URL.URL, true)
		if alreadyVisited {
			// Release the token before skipping
			<-c.sem
			continue
		}

		// Check depth and skip if it exceeds maximum depth
		if URL.Depth >= 1 {
			<-c.sem
			return
		}

		pageCtx, pageCancel := context.WithTimeout(context.Background(), time.Minute*time.Duration(c.timeout))
		fmt.Println("Crawling URL", URL.URL, "Depth:", URL.Depth)
		resp, err := http.Get(URL.URL)
		if err != nil {
			fmt.Println("Fetching error:", err)
			resp.Body.Close()
			pageCancel()
			<-c.sem
			continue
		}
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Reading body: ", err)
			resp.Body.Close()
			pageCancel()
			<-c.sem
			continue
		}
		resp.Body.Close()

		err = extracter.DownloadImages(pageCtx, c.imagesFolderName, c.db, b, URL.URL)
		if err != nil {
			fmt.Println("DownloadImages Err", err)
		}

		links, err := extracter.ExtractLinks(pageCtx, b, URL.URL)
		if err != nil {
			fmt.Println("ExtractLinks Err", err)
			pageCancel()
			<-c.sem
			continue
		}

		// Enqueue extracted links for further crawling with increased depth
		for _, link := range links {
			c.urlQueue.Enqueue(link, URL.Depth+1)

			// Start a new goroutine for each new URL
			c.wg.Add(1)
			go c.Crawl()
		}

		pageCancel()
		// Release the token after finishing the work
		<-c.sem
	}
}
