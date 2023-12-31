package crawler

import (
	"awesomeProject1/apps/imagecrawler/extracter"
	"awesomeProject1/db"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

const maxDepth = 1

type job struct {
	URL   string
	Depth int
}

type Crawler struct {
	jobs             chan job
	workers          int
	db               db.DB
	timeout          int
	imagesFolderName string
	visited          *sync.Map
	wg               *sync.WaitGroup
}

func NewCrawler(db db.DB, workers int, timeout int, imagesFolderName string) *Crawler {
	return &Crawler{
		jobs:             make(chan job),
		workers:          workers,
		db:               db,
		timeout:          timeout,
		imagesFolderName: imagesFolderName,
		visited:          &sync.Map{},
		wg:               &sync.WaitGroup{},
	}
}

func (c *Crawler) enqueue(j job) {
	c.wg.Add(1)
	c.jobs <- j
}

func (c *Crawler) Start(startingLinks []string) {
	if err := os.MkdirAll("apps/imagecrawler/"+c.imagesFolderName, os.ModePerm); err != nil {
		fmt.Printf("Error creating image directory: %v\n", err)
		os.Exit(1)
	}

	for i := 0; i < c.workers; i++ {
		go func() {
			for j := range c.jobs {
				c.runJob(j)
				c.wg.Done()
			}
		}()
	}

	for _, link := range startingLinks {
		c.enqueue(job{URL: link, Depth: 0})
	}
	c.wg.Wait()
	close(c.jobs)
}

func (c *Crawler) runJob(j job) {
	_, alreadyVisited := c.visited.LoadOrStore(j.URL, true)
	if alreadyVisited {
		return
	}

	pageCtx, pageCancel := context.WithTimeout(context.Background(), time.Minute*time.Duration(c.timeout))
	fmt.Println("Crawling URL", j.URL, "Depth:", j.Depth)
	resp, err := http.Get(j.URL)
	if err != nil {
		fmt.Println("Fetching error:", err)
		resp.Body.Close()
		pageCancel()
		return
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Reading body: ", err)
		resp.Body.Close()
		pageCancel()
		return
	}
	resp.Body.Close()

	err = extracter.DownloadImages(pageCtx, c.imagesFolderName, c.db, b, j.URL)
	if err != nil {
		fmt.Println("DownloadImages Err", err)
	}

	if j.Depth >= maxDepth {
		pageCancel()
		return
	}

	links, err := extracter.ExtractLinks(pageCtx, b, j.URL)
	if err != nil {
		fmt.Println("ExtractLinks Err", err)
		pageCancel()
		return
	}

	for _, link := range links {
		c.enqueue(job{URL: link, Depth: j.Depth + 1})
	}

	pageCancel()
}
