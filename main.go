package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"comic-crawler/service"

	"github.com/gocolly/colly"
	"github.com/joho/godotenv"
)

func init() {
	folders := []string{"out", "raw"}
	for _, folder := range folders {
		if err := service.OverwriteFolder(folder); err != nil {
			fmt.Println(err)
			return
		}
	}

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {
	domain := os.Getenv("DOMAIN")
	comicId := os.Getenv("COMIC_ID")

	fmt.Println("Trying to get list of chapters...")
	chapters, err := service.LoadChapters()
	if err != nil {
		fmt.Println(err)
		return
	}

	// Init crawler
	fmt.Println("Starting crawler...")
	c := colly.NewCollector(
		colly.AllowedDomains(domain, "www."+domain),
	)

	// Init downloader
	fmt.Println("Starting downloader...")
	worker := 1
	downloadWorker := os.Getenv("DOWNLOAD_WORKER") // Number of workers to download images concurrently
	if downloadWorker != "" {
		worker, _ = strconv.Atoi(downloadWorker)
	}
	fmt.Printf("Number of workers: %d\n", worker)

	fmt.Println("-----------------------------------")
	for _, chapter := range chapters {
		if ok, err := skipChapter(comicId, chapter); err != nil {
			fmt.Println(err)
			continue
		} else if ok {
			fmt.Printf("Chapter %s already exists, skipping...\n", chapter.Name)
			continue
		}

		fmt.Printf("Crawling chap %v...\n", chapter.Name)
		urls := service.CrawlImg(c, domain, chapter.Url)
		if len(urls) == 0 {
			fmt.Println("No images found")
			continue
		}

		var wg sync.WaitGroup
		jobs := make(chan URL) // Channel for sending URLs to download jobs
		wg.Add(len(urls))      // Set the wait group size to the number of URLs
		folder := fmt.Sprintf("out/%s/%s/", comicId, chapter.Name)
		fmt.Println("Creating folder ", folder)
		if err := service.OverwriteFolder(folder); err != nil {
			fmt.Println(err)
			continue
		}

		fmt.Println("Downloading images...")
		for i := 0; i < worker; i++ {
			workerId := i + 1
			go func(workerId int) {
				for job := range jobs {
					if job.Url != "" {
						// Download image
						filepath := strings.Split(job.Url, "/")
						dest := fmt.Sprintf("%s%d.%s", folder, job.Id, strings.Split(filepath[len(filepath)-1], ".")[1])
						if err := service.DownloadImage(workerId, job.Url, dest); err != nil {
							fmt.Println(err)
						}
					}
					wg.Done()
				}
			}(workerId)
		}

		// Send URLs to download pool
		for i, url := range urls {
			jobs <- URL{
				Id:  i + 1,
				Url: url,
			}
		}

		// Close pool channel to signal no more URLs to download
		close(jobs)

		// Wait for all download jobs to finish
		wg.Wait()

		sleep := 2 * time.Second
		fmt.Printf("Sleeping for %v\n", sleep)
		time.Sleep(sleep)
	}

	fmt.Println("Done!")
}

type URL struct {
	Id  int
	Url string
}

func skipChapter(comicId string, chapter service.Chapter) (bool, error) {
	folder := fmt.Sprintf("out/%s/%s/", comicId, chapter.Name)
	if ok, err := service.IsFolderExist(folder); err != nil {
		return false, err
	} else if ok {
		return true, nil
	}
	return false, nil
}
