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
	"comic-crawler/service/crawler"
	"comic-crawler/service/epub"

	"github.com/gocolly/colly"
	"github.com/joho/godotenv"
	"github.com/vukyn/kuery/query/v2"
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
	timeStart := time.Now()
	// crawl()
	convert()
	fmt.Printf("Done for %.2fs!\n", time.Since(timeStart).Seconds())
}

func crawl() {
	domain := os.Getenv("DOMAIN")
	comicId := getComicId(domain)

	// Init crawler
	fmt.Println("Starting crawler...")
	c := colly.NewCollector(
		colly.AllowedDomains(domain, "www."+domain),
	)

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		RandomDelay: 1 * time.Second,
	})

	fmt.Println("Trying to get list of chapters...")
	chapters, err := crawler.CrawlChapter(c, domain)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Init downloader
	fmt.Println("Starting downloader...")
	worker := 1
	downloadWorker := os.Getenv("DOWNLOAD_WORKER") // Number of workers to download images concurrently
	if downloadWorker != "" {
		worker, _ = strconv.Atoi(downloadWorker)
	}
	fmt.Printf("Number of workers: %d\n", worker)

	fmt.Println("-----------------------------------")

	isCrawlAll := os.Getenv("CRAWL_ALL")
	crawlChaptersEnv := os.Getenv("CRAWL_CHAPTERS")
	crawlChapters := make([]string, 0)
	if crawlChaptersEnv != "" {
		if strings.Contains(crawlChaptersEnv, ",") {
			crawlChapters = strings.Split(crawlChaptersEnv, ",")
		} else if strings.Contains(crawlChaptersEnv, "-") {
			crawlRange := strings.Split(crawlChaptersEnv, "-")
			start, _ := strconv.Atoi(crawlRange[0])
			end, _ := strconv.Atoi(crawlRange[1])
			if start > end {
				fmt.Println("Invalid range")
				return
			}
			for i := start; i <= end; i++ {
				crawlChapters = append(crawlChapters, fmt.Sprint(i))
			}
		} else {
			crawlChapters = append(crawlChapters, crawlChaptersEnv)
		}
	}

	for _, chapter := range chapters {
		if isCrawlAll == "" || isCrawlAll == "false" {
			if isAny := query.AnyFunc(crawlChapters, func(i string) bool {
				return "Chapter "+i == chapter.Name || "Chương "+i == chapter.Name || i == chapter.Name
			}); !isAny {
				fmt.Printf("Skipping chapter %s...\n", chapter.Name)
				continue
			}
		}
		if ok, err := skipChapter(domain, comicId, chapter.Name); err != nil {
			fmt.Println(err)
			continue
		} else if ok {
			fmt.Printf("Chapter %s already exists, skipping...\n", chapter.Name)
			continue
		}

		fmt.Printf("Crawling chap %v...\n", chapter.Name)
		urls := crawler.CrawlImg(c, domain, chapter.Url)
		if len(urls) == 0 {
			fmt.Println("No images found")
			continue
		}

		var wg sync.WaitGroup
		jobs := make(chan URL) // Channel for sending URLs to download jobs
		wg.Add(len(urls))      // Set the wait group size to the number of URLs
		folder := getFolderPath(domain, comicId, chapter.Name)
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
						if err := service.DownloadImage(workerId, job.Url, domain, dest); err != nil {
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

		// Close jobs channel to signal no more URLs to download
		close(jobs)

		// Wait for all download jobs to finish
		wg.Wait()

		sleep()
	}
}

func convert() {
	domain := os.Getenv("DOMAIN")
	convertFormat := os.Getenv("CONVERT_FORMAT")
	convertComicId := os.Getenv("CONVERT_COMIC_ID")

	comicPath := fmt.Sprintf("out/%s/%s", getWebsiteName(domain), convertComicId)
	files, err := os.ReadDir(comicPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Comic not found")
			return
		}
		fmt.Println(err)
		return
	}

	if convertFormat != "" {
		fmt.Println("Converting...")
		convertList := strings.Split(convertFormat, ",")

		titleName := os.Getenv("TITLE")
		if titleName == "" {
			titleName = "Title"
		}

		author := os.Getenv("AUTHOR")
		if author == "" {
			author = "Unknown"
		}
		
		wg := sync.WaitGroup{}
		for _, format := range convertList {
			switch strings.ToLower(format) {
			case "pdf":
			case "epub":
				wg.Add(len(files))
				for i := range files {
					go func(i int) {
						if !files[i].IsDir() {
							return
						}
						chapterPath := fmt.Sprintf("out/%s/%s/%s", getWebsiteName(domain), convertComicId, files[i].Name())
						epubOpt := epub.EpubOption{
							Title:  fmt.Sprintf("%s - %s", titleName, files[i].Name()),
							Author: author,
							Cover:  os.Getenv("COVER"),
						}
						if err := epub.ImagesToEPUB(chapterPath, comicPath, files[i].Name(), epubOpt); err != nil {
							fmt.Println(err)
						}
						fmt.Println("Converted ", files[i].Name(), " to EPUB...")
						wg.Done()
					}(i)
				}
				wg.Wait()
			default:
			}
		}
	}

}

type URL struct {
	Id  int
	Url string
}

func skipChapter(domain, comicId, chapterName string) (bool, error) {
	folder := getFolderPath(domain, comicId, chapterName)
	if ok, err := service.IsFolderExist(folder); err != nil {
		return false, err
	} else if ok {
		return true, nil
	}
	return false, nil
}

func getComicId(domain string) string {
	var comicId = map[string]string{
		os.Getenv("NETTRUYEN_DOMAIN"): os.Getenv("NETTRUYEN_COMIC_ID"),
		os.Getenv("QQTRUYEN_DOMAIN"):  os.Getenv("QQTRUYEN_COMIC_ID"),
	}
	return comicId[domain]
}

func getFolderPath(domain, comicId, chapterName string) string {
	return fmt.Sprintf("out/%s/%s/%s/", getWebsiteName(domain), comicId, chapterName)
}

func getWebsiteName(domain string) string {
	var websiteName = map[string]string{
		os.Getenv("NETTRUYEN_DOMAIN"): "nettruyen",
		os.Getenv("QQTRUYEN_DOMAIN"):  "qqtruyen",
	}
	return websiteName[domain]
}

func sleep() {
	sleep := os.Getenv("SLEEP")
	sleepTime, _ := strconv.Atoi(sleep)
	if sleepTime == 0 {
		sleepTime = 2000
	}
	fmt.Printf("Sleeping for %vms\n", sleepTime)
	time.Sleep(time.Duration(sleepTime) * time.Millisecond)
}
