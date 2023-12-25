package main

import (
	"awesomeProject1/apps/httpserver"
	"awesomeProject1/apps/imagecrawler"
	"fmt"
	"os"
)

func main() {
	fmt.Println("Menu:")
	fmt.Println("1. Crawl images")
	fmt.Println("2. Serve HTTP server")

	fmt.Print("Enter your choice: ")
	var choice int
	_, err := fmt.Scan(&choice)

	if err != nil {
		fmt.Println("Invalid input. Please enter a number.")
		os.Exit(0)
	}

	switch choice {
	case 1:
		imagecrawler.RunCrawler()
	case 2:
		httpserver.RunServer()
	default:
		fmt.Println("Invalid choice. Please enter a valid option.")
		os.Exit(0)
	}
}
