package crawler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/gocolly/colly"
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
		os.Getenv("NETTRUYEN_DOMAIN"): nettruyenChapterCallback,
		os.Getenv("QQTRUYEN_DOMAIN"):  qqtruyenChapterCallback,
	}
	callback, ok := crawler[domain]
	if !ok {
		fmt.Println("Domain not supported")
		return nil, fmt.Errorf("Domain not supported")
	}
	return callback(c)
}

func nettruyenChapterCallback(_ *colly.Collector) ([]Chapter, error) {
	res, err := makeGet(os.Getenv("NETTRUYEN_COMIC_QUERY"))
	if err != nil {
		return nil, err
	}

	var chapterResponse ChapterResponse
	err = json.Unmarshal(res, &chapterResponse)
	if err != nil {
		fmt.Println("Error unmarshalling data:", err)
		return nil, err
	}

	if !chapterResponse.Success {
		fmt.Println("Error loading chapters:", string(res))
		return nil, err
	}

	for _, chapter := range chapterResponse.Chapters {
		fmt.Printf("Link found: %s\n", chapter.Url)
	}
	fmt.Printf("Total chapter found (%d)\n", len(chapterResponse.Chapters))

	return chapterResponse.Chapters, nil
}

func qqtruyenChapterCallback(c *colly.Collector) ([]Chapter, error) {
	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	chapters := make([]Chapter, 0)
	c.OnHTML("div.works-chapter-list", func(e *colly.HTMLElement) {
		e.ForEach("div.works-chapter-item", func(i int, chapterItem *colly.HTMLElement) {
			chapterItem.ForEach("a", func(_ int, e1 *colly.HTMLElement) {
				link, _ := url.Parse(e1.Attr("href"))
				fmt.Printf("Link found: %s\n", link.Path)
				chapters = append(chapters, Chapter{
					Id:   i + 1,
					Name: e1.Text,
					Url:  link.Path,
				})
			})
		})
	})

	// Start scraping
	if err := c.Visit(os.Getenv("QQTRUYEN_COMIC_QUERY")); err != nil {
		fmt.Println(err)
	}
	fmt.Printf("Total chapter found (%d)\n", len(chapters))

	return chapters, nil
}

func makeGet(url string) ([]byte, error) {
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return nil, err
	}

	data, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return nil, err
	}
	defer data.Body.Close()

	res, err := io.ReadAll(data.Body)
	if err != nil {
		fmt.Println("Error reading data:", err)
		return nil, err
	}

	return res, nil
}
