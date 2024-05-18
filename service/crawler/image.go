package crawler

import (
	"comic-crawler/env"
	"net/url"

	"github.com/gocolly/colly"
	"github.com/vukyn/kuery/log"
	"github.com/vukyn/kuery/query/v2"
)

type Collector struct {
	Url []string
}

func CrawlImg(c *colly.Collector, domain, url string) []string {
	var crawler = map[string]func(*colly.Collector, *Collector){
		env.NettruyenDomain: nettruyenImgCallback,
		env.QqtruyenDomain:  qqtruyenImgCallback,
	}
	callback, ok := crawler[domain]
	if !ok {
		log.Errorf("Domain not supported: %s", domain)
		return nil
	}

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		log.Infof("Visiting %s", r.URL.String())
	})

	// Callback
	imgCollector := &Collector{}
	callback(c, imgCollector)

	// Start scraping
	url = "https://" + domain + url
	if err := c.Visit(url); err != nil {
		log.Errorf("Error visiting: %v", err)
		return nil
	}
	log.Infof("Total link found (%d)", len(imgCollector.Url))

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
			log.Infof("Link found: %s", link.String())
			imgCollector.Url = append(imgCollector.Url, link.String())
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
			log.Infof("Link found: %s", link.String())
			imgCollector.Url = append(imgCollector.Url, link.String())
		})
	})
}
