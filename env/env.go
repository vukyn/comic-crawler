package env

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

const (
	DEFAULT_DOMAIN                  = ""
	DEFAULT_COMIC_ID                = 0
	DEFAULT_NETTRUYEN_DOMAIN        = "nettruyendie.com"
	DEFAULT_NETTRUYEN_REFERER       = ""
	DEFAULT_NETTRUYEN_CHAPTER_QUERY = "Comic/Services/ComicService.asmx/ProcessChapterList"
	DEFAULT_QQTRUYEN_DOMAIN         = "truyenqqviet.com"
	DEFAULT_QQTRUYEN_REFERER        = "https://truyenqqviet.com/"
	DEFAULT_QQTRUYEN_CHAPTER_QUERY  = ""
	DEFAULT_CRAWL_ALL               = false
	DEFAULT_CRAWL_CHAPTERS          = ""
	DEFAULT_CRAWL_WORKER            = 8
	DEFAULT_DOWNLOAD_WORKER         = 1
	DEFAULT_SLEEP                   = 2000
	DEFAULT_COVER                   = ""
	DEFAULT_TITLE                   = "Title"
	DEFAULT_AUTHOR                  = "Unknown"
	DEFAULT_CONVERT_FORMAT          = "EPUB"
	DEFAULT_CONVERT_COMIC_ID        = ""
)

var (
	Domain                string
	ComicId               int
	NettruyenDomain       string
	NettruyenReferer      string
	NettruyenChapterQuery string
	QqtruyenDomain        string
	QqtruyenReferer       string
	QqtruyenChapterQuery  string
	CrawlAll              bool
	CrawlChapters         string
	CrawlWorker           int
	DownloadWorker        int
	Sleep                 int
	Cover                 string
	Title                 string
	Author                string
	ConvertFormat         string
)

func Init() error {
	env, err := loadEnv()
	if err != nil {
		return err
	}

	if domain, ok := env["DOMAIN"]; ok {
		Domain = domain
	} else {
		Domain = DEFAULT_DOMAIN
	}

	if comicId, ok := env["COMIC_ID"]; ok {
		ComicId, _ = strconv.Atoi(comicId)
	} else {
		ComicId = DEFAULT_COMIC_ID
	}

	if nettruyenDomain, ok := env["NETTRUYEN_DOMAIN"]; ok {
		NettruyenDomain = nettruyenDomain
	} else {
		NettruyenDomain = DEFAULT_NETTRUYEN_DOMAIN
	}

	if nettruyenReferer, ok := env["NETTRUYEN_REFERER"]; ok {
		NettruyenReferer = nettruyenReferer
	} else {
		NettruyenReferer = DEFAULT_NETTRUYEN_REFERER
	}

	if nettruyenChapterQuery, ok := env["NETTRUYEN_CHAPTER_QUERY"]; ok {
		NettruyenChapterQuery = nettruyenChapterQuery
	} else {
		NettruyenChapterQuery = DEFAULT_NETTRUYEN_CHAPTER_QUERY
	}

	if qqtruyenDomain, ok := env["QQTRUYEN_DOMAIN"]; ok {
		QqtruyenDomain = qqtruyenDomain
	} else {
		QqtruyenDomain = DEFAULT_QQTRUYEN_DOMAIN
	}

	if qqtruyenReferer, ok := env["QQTRUYEN_REFERER"]; ok {
		QqtruyenReferer = qqtruyenReferer
	} else {
		QqtruyenReferer = DEFAULT_QQTRUYEN_REFERER
	}

	if qqtruyenChapterQuery, ok := env["QQTRUYEN_CHAPTER_QUERY"]; ok {
		QqtruyenChapterQuery = qqtruyenChapterQuery
	} else {
		QqtruyenChapterQuery = DEFAULT_QQTRUYEN_CHAPTER_QUERY
	}

	if crawlAll, ok := env["CRAWL_ALL"]; ok {
		CrawlAll, _ = strconv.ParseBool(crawlAll)
	} else {
		CrawlAll = DEFAULT_CRAWL_ALL
	}

	if crawlChapters, ok := env["CRAWL_CHAPTERS"]; ok {
		CrawlChapters = crawlChapters
	} else {
		CrawlChapters = DEFAULT_CRAWL_CHAPTERS
	}

	if crawlWorker, ok := env["CRAWL_WORKER"]; ok {
		CrawlWorker, _ = strconv.Atoi(crawlWorker)
	} else {
		CrawlWorker = DEFAULT_CRAWL_WORKER
	}

	if downloadWorker, ok := env["DOWNLOAD_WORKER"]; ok {
		DownloadWorker, _ = strconv.Atoi(downloadWorker)
	} else {
		DownloadWorker = DEFAULT_DOWNLOAD_WORKER
	}

	if sleep, ok := env["SLEEP"]; ok {
		Sleep, _ = strconv.Atoi(sleep)
	} else {
		Sleep = DEFAULT_SLEEP
	}

	if cover, ok := env["COVER"]; ok {
		Cover = cover
	} else {
		Cover = DEFAULT_COVER
	}

	if title, ok := env["TITLE"]; ok {
		Title = title
	} else {
		Title = DEFAULT_TITLE
	}

	if author, ok := env["AUTHOR"]; ok {
		Author = author
	} else {
		Author = DEFAULT_AUTHOR
	}

	if convertFormat, ok := env["CONVERT_FORMAT"]; ok {
		ConvertFormat = convertFormat
	} else {
		ConvertFormat = DEFAULT_CONVERT_FORMAT
	}

	return nil
}

func loadEnv() (map[string]string, error) {
	env, err := os.ReadFile(".env")
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]string{}, nil
		}
		return nil, err
	}
	envMap, err := godotenv.UnmarshalBytes(env)
	if err != nil {
		return nil, err
	}

	return envMap, nil
}
