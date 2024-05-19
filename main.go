package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"comic-crawler/env"
	"comic-crawler/service"
	"comic-crawler/service/crawler"
	"comic-crawler/service/downloader"
	"comic-crawler/service/epub"

	"github.com/gocolly/colly"
	"github.com/vukyn/kuery/file"
	"github.com/vukyn/kuery/log"
	"github.com/vukyn/kuery/query/v2"
)

func init() {
	// Init folders
	folders := []string{"out", "raw"}
	for _, folder := range folders {
		if err := file.CreateFilePath(folder); err != nil {
			log.Errorf("Failed to create folder %s: %v", folder, err)
			return
		}
	}

	// Init log
	log.SetPrettyLog()

	// Load env
	if err := env.Init(); err != nil {
		log.Errorf("Failed to load env: %v", err)
		return
	}
}

func main() {
	timeStart := time.Now()
	// crawl()
	convert()
	log.Infof("Done for %.2fs!", time.Since(timeStart).Seconds())
}

func crawl() {
	domain := env.Domain
	comicId := env.ComicId

	// Init crawler
	log.Infof("Starting crawler...")
	c := colly.NewCollector(
		colly.AllowedDomains(domain, "www."+domain),
	)

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		RandomDelay: 1 * time.Second,
	})

	log.Infof("Trying to get list of chapters...")
	chapters, err := crawler.CrawlChapter(c, domain)
	if err != nil {
		log.Errorf("Failed to get list of chapters: %v", err)
		return
	}

	// Init downloader
	log.Infof("Starting downloader...")
	downloadWorker := env.DownloadWorker
	log.Infof("Number of download workers: %d", downloadWorker)

	fmt.Println("-----------------------------------")

	isCrawlAll := env.CrawlAll
	crawlChaptersEnv := env.CrawlChapters
	crawlChapters := make([]string, 0)
	if crawlChaptersEnv != "" {
		if strings.Contains(crawlChaptersEnv, ",") {
			crawlChapters = strings.Split(crawlChaptersEnv, ",")
		} else if strings.Contains(crawlChaptersEnv, "-") {
			crawlRange := strings.Split(crawlChaptersEnv, "-")
			start, _ := strconv.Atoi(crawlRange[0])
			end, _ := strconv.Atoi(crawlRange[1])
			if start > end {
				log.Errorf("Invalid range: %s", crawlChaptersEnv)
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
		if !isCrawlAll {
			if isAny := query.AnyFunc(crawlChapters, func(i string) bool {
				return "Chapter "+i == chapter.Name || "Chương "+i == chapter.Name || i == chapter.Name
			}); !isAny {
				log.Warnf("Skipping chapter %s...", chapter.Name)
				continue
			}
		}
		if ok, err := skipChapter(domain, chapter.Name, comicId); err != nil {
			log.Errorf("Failed to skip chapter %s: %v", chapter.Name, err)
			continue
		} else if ok {
			log.Warnf("Chapter %s already exists, skipping...", chapter.Name)
			continue
		}

		log.Infof("Crawling chap %v...", chapter.Name)
		urls := crawler.CrawlImg(c, domain, chapter.Url)
		if len(urls) == 0 {
			log.Errorf("No images found")
			continue
		}

		var wg sync.WaitGroup
		jobs := make(chan URL) // Channel for sending URLs to download jobs
		wg.Add(len(urls))      // Set the wait group size to the number of URLs
		folder := getFolderPath(domain, chapter.Name, comicId)
		log.Infof("Creating folder %s", folder)
		if err := service.OverwriteFolder(folder); err != nil {
			log.Errorf("Failed to overwrite folder %s: %v", folder, err)
			continue
		}

		log.Infof("Downloading images...")
		for i := 0; i < downloadWorker; i++ {
			workerId := i + 1
			go func(workerId int) {
				for job := range jobs {
					if job.Url != "" {
						// Download image
						dest := fmt.Sprintf("%s%d.jpg", folder, job.Id)
						if err := downloader.DownloadImg(workerId, job.Url, domain, dest); err != nil {
							log.Errorf("Failed to download image %s: %v", job.Url, err)
						} else {
							log.Infof("Downloaded %s", job.Url)
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
	domain := env.Domain
	comicId := env.ComicId
	convertFormat := env.ConvertFormat

	comicPath := fmt.Sprintf("out/%s/%d", getWebsiteName(domain), comicId)
	files, err := os.ReadDir(comicPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Errorf("Comic not found")
			return
		}
		log.Errorf("Failed to read comic folder: %v", err)
		return
	}

	if convertFormat != "" {
		log.Infof("Converting...")
		convertList := strings.Split(convertFormat, ",")

		wg := sync.WaitGroup{}
		for _, format := range convertList {
			switch strings.ToUpper(format) {
			case "PDF":
			case "EPUB":
				wg.Add(len(files))
				for i := range files {
					go func(i int) {
						if !validFolderChapter(files[i]) {
							wg.Done()
							return
						}
						chapterPath := fmt.Sprintf("out/%s/%d/%s", getWebsiteName(domain), comicId, files[i].Name())
						cover := env.Cover
						if cover == "" {
							cover = randomCover()
						}
						epubOpt := epub.EpubOption{
							Title:  fmt.Sprintf("%s - %s", env.Title, files[i].Name()),
							Author: env.Author,
							Cover:  cover,
						}
						if err := epub.ImagesToEPUB(chapterPath, comicPath, files[i].Name(), epubOpt); err != nil {
							log.Errorf("Failed to convert %s: %v", files[i].Name(), err)
							wg.Done()
							return
						}
						log.Infof("Converted %s to EPUB", files[i].Name())
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

func skipChapter(domain, chapterName string, comicId int) (bool, error) {
	folder := getFolderPath(domain, chapterName, comicId)
	if ok, err := service.IsFolderExist(folder); err != nil {
		return false, err
	} else if ok {
		return true, nil
	}
	return false, nil
}

func getFolderPath(domain, chapterName string, comicId int) string {
	return fmt.Sprintf("out/%s/%d/%s/", getWebsiteName(domain), comicId, chapterName)
}

func validFolderChapter(f os.DirEntry) bool {
	isFile := !f.IsDir()
	if isFile ||
		f.Name() == "epub" ||
		f.Name() == "pdf" ||
		f.Name() == "temp" {
		return false
	}
	return true
}

func randomCover() string {
	folderPath := "assets/default"
	files, err := os.ReadDir(folderPath)
	if err != nil {
		log.Errorf("Failed to read default cover folder: %v", err)
		return ""
	}
	imgs := make([]string, 0)
	for _, file := range files {
		if !file.IsDir() &&
			strings.Contains(file.Name(), ".jpg") &&
			strings.Contains(file.Name(), "default_cover_") {
			imgs = append(imgs, fmt.Sprintf("%s/%s", folderPath, file.Name()))
		}
	}

	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)
	return imgs[r.Intn(len(imgs))]
}

func getWebsiteName(domain string) string {
	var websiteName = map[string]string{
		env.NettruyenDomain: "nettruyen",
		env.QqtruyenDomain:  "qqtruyen",
	}
	return websiteName[domain]
}

func sleep() {
	log.Infof("Sleeping for %vms", env.Sleep)
	time.Sleep(time.Duration(env.Sleep) * time.Millisecond)
}
