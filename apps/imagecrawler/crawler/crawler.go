package crawler

import (
	"awesomeProject1/apps/imagecrawler/extracter"
	"awesomeProject1/db"
	"awesomeProject1/logger"
	"context"
	"io"
	"net/http"
	"sync"
	"time"
)

var (
	log      = logger.New()
	maxDepth = 0
)

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
	downloaded       *sync.Map
	wg               *sync.WaitGroup
}

func NewCrawler(db db.DB, workers int, timeout int, imagesFolderName string) *Crawler {
	return &Crawler{
		jobs:             make(chan job, 1),
		workers:          workers,
		db:               db,
		timeout:          timeout,
		imagesFolderName: imagesFolderName,
		visited:          &sync.Map{},
		downloaded:       &sync.Map{},
		wg:               &sync.WaitGroup{},
	}
}

func (c *Crawler) Start(startingLinks []string) {
	for i := 0; i < c.workers; i++ {
		go func() {
			for j := range c.jobs {
				c.runJob(j)
				c.wg.Done()
			}
		}()
	}

	for _, link := range startingLinks {
		c.wg.Add(1)
		c.jobs <- job{URL: link, Depth: 0}
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
	log.Infoln("Crawling URL", j.URL, "Depth:", j.Depth)
	resp, err := http.Get(j.URL)
	if err != nil {
		log.Infof("Fetching error: %s", err)
		resp.Body.Close()
		pageCancel()
		return
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Infof("Reading body error: %s", err)
		resp.Body.Close()
		pageCancel()
		return
	}
	resp.Body.Close()

	err = extracter.DownloadImages(pageCtx, c.imagesFolderName, c.db, b, j.URL, c.downloaded)
	if err != nil {
		log.Infof("DownloadImages error: %s", err)
	}

	if j.Depth >= maxDepth {
		pageCancel()
		return
	}

	links, err := extracter.ExtractLinks(pageCtx, b, j.URL)
	if err != nil {
		log.Infof("ExtractLinks error: %s", err)
		pageCancel()
		return
	}

	for _, link := range links {
		c.wg.Add(1)
		go func(link string) {
			c.jobs <- job{URL: link, Depth: j.Depth + 1}
		}(link)
	}
	pageCancel()
}
