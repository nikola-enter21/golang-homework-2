package imagecrawler

import (
	"awesomeProject1/apps/imagecrawler/crawler"
	"awesomeProject1/apps/imagecrawler/queue"
	"awesomeProject1/config"
	"awesomeProject1/db/postgres"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
)

func RunCrawler() {
	cfg := config.MustParseConfig()
	var wg sync.WaitGroup
	urlQueue := &queue.URLQueue{}
	visited := &sync.Map{} // Use a sync.Map to store visited URLs

	if err := os.MkdirAll("apps/imagecrawler/"+cfg.ImagesFolderName, os.ModePerm); err != nil {
		fmt.Printf("Error creating image directory: %v\n", err)
		os.Exit(1)
	}

	// Command line flags
	urls := flag.String("urls", "https://www.slickerhq.com", "Comma-separated list of URLs to start crawling from")
	timeoutInMinutes := flag.Uint("page-timeout", 2, "Crawling timeout per page duration in minutes")
	maxWorkers := flag.Int("max-workers", 100, "Maximum number of goroutines")
	flag.Parse()

	if *urls == "" {
		fmt.Println("Please provide at least one URL to start crawling from")
		os.Exit(1)
	}
	startURLs := strings.Split(*urls, ",")
	for _, url := range startURLs {
		urlQueue.Enqueue(url, 0)
	}

	db, err := postgres.NewPostgresDatabase(cfg.PostgresDSN)
	if err != nil {
		fmt.Printf("Error connecting to the database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Use a buffered channel as a semaphore to limit the number of goroutines
	sem := make(chan struct{}, *maxWorkers)
	c := crawler.NewCrawler(
		urlQueue,
		db,
		*timeoutInMinutes,
		cfg.ImagesFolderName,
		visited,
		sem,
		&wg,
	)

	for i := 0; i < min(*maxWorkers, len(startURLs)); i++ {
		wg.Add(1)
		go c.Crawl()
	}

	wg.Wait()
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}
