package service

import (
	"fmt"
	"os"

	"net/url"

	"github.com/gocolly/colly"
	"github.com/vukyn/kuery/query/v2"
)

type Collector struct {
	Url []string
}

func CrawlImg(c *colly.Collector, domain, url string) []string {

	var crawler = map[string]func(*colly.Collector, *Collector){
		os.Getenv("NETTRUYEN_DOMAIN"): nettruyenCallback,
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
		fmt.Println(err)
	}
	fmt.Printf("Total link found (%d)\n", len(imgCollector.Url))

	return query.Map(imgCollector.Url, func(c string) string {
		return c
	})
}

func nettruyenCallback(c *colly.Collector, imgCollector *Collector) {
	c.OnHTML("div.page-chapter", func(e *colly.HTMLElement) {
		e.ForEach("img.lozad", func(_ int, e1 *colly.HTMLElement) {
			link := e1.Attr("src")
			if link == "" {
				link = e1.Attr("data-src")
			}
			link = processURL(link)
			fmt.Printf("Link found: %s\n", link)
			imgCollector.Url = append(imgCollector.Url, link)
		})
	})
}

func processURL(path string) string {
	u, _ := url.Parse(path)
	return u.Host + u.Path
}
