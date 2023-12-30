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
	visited := &sync.Map{}

	if err := os.MkdirAll("apps/imagecrawler/"+cfg.ImagesFolderName, os.ModePerm); err != nil {
		fmt.Printf("Error creating image directory: %v\n", err)
		os.Exit(1)
	}

	// Command line flags
	urls := flag.String("urls", "https://www.slickerhq.com,https://www.ycombinator.com", "Comma-separated list of URLs to start crawling from")
	timeoutInMinutes := flag.Uint("page-timeout", 100, "Crawling timeout per page duration in minutes")
	maxWorkers := flag.Int("max-workers", 1000, "Maximum number of goroutines")
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

	c := crawler.NewCrawler(
		urlQueue,
		db,
		*timeoutInMinutes,
		cfg.ImagesFolderName,
		visited,
		&wg,
	)

	wg.Add(*maxWorkers)
	for i := 0; i < *maxWorkers; i++ {
		go c.Crawl()
	}
	wg.Wait()
}
