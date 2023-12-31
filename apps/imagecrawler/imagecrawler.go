package imagecrawler

import (
	"awesomeProject1/apps/imagecrawler/crawler"
	"awesomeProject1/config"
	"awesomeProject1/db/postgres"
	"flag"
	"fmt"
	"os"
	"strings"
)

func RunCrawler() {
	cfg := config.MustParseConfig()

	// Command line flags
	urls := flag.String("urls", "https://www.slickerhq.com,https://www.ycombinator.com", "Comma-separated list of URLs to start crawling from")
	timeoutInMinutes := flag.Int("page-timeout", 100, "Crawling timeout per page duration in minutes")
	maxWorkers := flag.Int("max-workers", 1000, "Maximum number of goroutines")
	flag.Parse()

	if *urls == "" {
		fmt.Println("Please provide at least one URL to start crawling from")
		os.Exit(1)
	}
	startURLs := strings.Split(*urls, ",")

	db, err := postgres.NewPostgresDatabase(cfg.PostgresDSN)
	if err != nil {
		fmt.Printf("Error connecting to the database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	c := crawler.NewCrawler(
		db,
		*maxWorkers,
		*timeoutInMinutes,
		cfg.ImagesFolderName,
	)

	c.Start(startURLs)
}
