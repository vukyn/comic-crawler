package crawler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"comic-crawler/env"

	"github.com/gocolly/colly"
	"github.com/vukyn/kuery/log"
)

type ChapterResponse struct {
	Success  bool      `json:"success"`
	Chapters []Chapter `json:"chapters"`
}

type Chapter struct {
	Id   int    `json:"chapterId"`
	Name string `json:"name"`
	Url  string `json:"url"`
}

func CrawlChapter(c *colly.Collector, domain string) ([]Chapter, error) {
	var crawler = map[string]func(*colly.Collector) ([]Chapter, error){
		env.NettruyenDomain: nettruyenChapterCallback,
		env.QqtruyenDomain:  qqtruyenChapterCallback,
	}
	callback, ok := crawler[domain]
	if !ok {
		log.Errorf("Domain not supported: %s", domain)
		return nil, fmt.Errorf("Domain not supported")
	}
	return callback(c)
}

func nettruyenChapterCallback(_ *colly.Collector) ([]Chapter, error) {
	url := fmt.Sprintf("https://www.%s/%s?comicId=%d", env.NettruyenDomain, env.NettruyenChapterQuery, env.ComicId)
	res, err := makeGet(url)
	if err != nil {
		return nil, err
	}

	var chapterResponse ChapterResponse
	err = json.Unmarshal(res, &chapterResponse)
	if err != nil {
		log.Errorf("Error unmarshalling response: %v", err)
		return nil, err
	}

	if !chapterResponse.Success {
		log.Errorf("Error getting chapters: %v", err)
		return nil, err
	}

	for _, chapter := range chapterResponse.Chapters {
		log.Infof("Chapter found: (%s) - %s", chapter.Name, chapter.Url)
	}
	log.Infof("Total chapter found (%d)", len(chapterResponse.Chapters))

	return chapterResponse.Chapters, nil
}

func qqtruyenChapterCallback(c *colly.Collector) ([]Chapter, error) {
	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		log.Infof("Visiting %s", r.URL.String())
	})

	chapters := make([]Chapter, 0)
	c.OnHTML("div.works-chapter-list", func(e *colly.HTMLElement) {
		e.ForEach("div.works-chapter-item", func(i int, chapterItem *colly.HTMLElement) {
			chapterItem.ForEach("a", func(_ int, e1 *colly.HTMLElement) {
				link, _ := url.Parse(e1.Attr("href"))
				log.Infof("Chapter found: (%s) - %s", e1.Text, link.Path)
				chapters = append(chapters, Chapter{
					Id:   i + 1,
					Name: e1.Text,
					Url:  link.Path,
				})
			})
		})
	})

	// Start scraping
	if err := c.Visit(env.QqtruyenChapterQuery); err != nil {
		log.Errorf("Error visiting: %v", err)
		return nil, err
	}
	log.Infof("Total chapter found (%d)", len(chapters))

	return chapters, nil
}

func makeGet(url string) ([]byte, error) {
	ctx, cancel := context.WithDeadline(context.TODO(), time.Now().Add(2*time.Second))
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.Errorf("Error creating request: %v", err)
		return nil, err
	}

	data, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Errorf("Error making request: %v", err)
		return nil, err
	}
	defer data.Body.Close()

	res, err := io.ReadAll(data.Body)
	if err != nil {
		log.Errorf("Error reading response: %v", err)
		return nil, err
	}
	return res, nil
}
