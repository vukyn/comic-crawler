package crawler

import (
	"fmt"
	"net/url"
	"os"

	"github.com/gocolly/colly"
	"github.com/vukyn/kuery/query/v2"
)

type Collector struct {
	Url []string
}

func CrawlImg(c *colly.Collector, domain, url string) []string {
	var crawler = map[string]func(*colly.Collector, *Collector){
		os.Getenv("NETTRUYEN_DOMAIN"): nettruyenImgCallback,
		os.Getenv("QQTRUYEN_DOMAIN"):  qqtruyenImgCallback,
	}
	callback, ok := crawler[domain]
	if !ok {
		fmt.Println("Domain not supported")
		return nil
	}

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	// Callback
	imgCollector := &Collector{}
	callback(c, imgCollector)

	// Start scraping
	url = "https://" + domain + url
	if err := c.Visit(url); err != nil {
		fmt.Println("Colly scraper: ", err)
	}
	fmt.Printf("Total link found (%d)\n", len(imgCollector.Url))

	return query.Map(imgCollector.Url, func(c string) string {
		return c
	})
}

func nettruyenImgCallback(c *colly.Collector, imgCollector *Collector) {
	c.OnHTML("div.page-chapter", func(e *colly.HTMLElement) {
		e.ForEach("img.lozad", func(_ int, e1 *colly.HTMLElement) {
			src := e1.Attr("src")
			if src == "" {
				src = e1.Attr("data-src")
			}
			link, _ := url.Parse(src)
			fmt.Printf("Link found: %s\n", link.Host+link.Path)
			imgCollector.Url = append(imgCollector.Url, link.Host+link.Path)
		})
	})
}

func qqtruyenImgCallback(c *colly.Collector, imgCollector *Collector) {
	c.OnHTML("div.chapter_content", func(e *colly.HTMLElement) {
		e.ForEach("img.lazy", func(_ int, e1 *colly.HTMLElement) {
			src := e1.Attr("src")
			if src == "" {
				src = e1.Attr("data-src")
			}
			link, _ := url.Parse(src)
			link.RawQuery = ""
			fmt.Printf("Link found: %s\n", link.String())
			imgCollector.Url = append(imgCollector.Url, link.String())
		})
	})
}