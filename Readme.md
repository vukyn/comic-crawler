## Configuration:

Environment variables are passed before crawling the website. The following environment variables are required:

| Name                    | Default                                               | Description                                                                                                                               |
| ----------------------- | ----------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------- |
| DOMAIN                  |                                                       | (Required) Website domain currently working                                                                                               |
| COMIC_ID                |                                                       | (Required) Comic id used to crawl or convert chapter                                                                                      |
| NETTRUYEN_DOMAIN        | 'nettruyendie.com'                                    | Nettruyen domain                                                                                                                          |
| NETTRUYEN_CHAPTER_QUERY | 'Comic/Services/ComicService.asmx/ProcessChapterList' | Query used to crawl all chapters                                                                                                          |
| QQTRUYEN_DOMAIN         | 'truyenqqviet.com'                                    | qqtruyen domain                                                                                                                           |
| QQTRUYEN_REFERER        | 'https://truyenqqviet.com/'                           | qqtruyen referer                                                                                                                          |
| QQTRUYEN_CHAPTER_QUERY  |                                                       | (Required) Full query url for crawl all chapter                                                                                           |
| CRAWL_ALL               | 'TRUE'                                                | Crawl all or specific chapter                                                                                                             |
| CRAWL_CHAPTERS          |                                                       | (Required if CRAWL_ALL is false) Crawling all given chapters, for example: <br> - 1,2,3,4 (delimiter by comma)<br> - 1-10 (range from-to) |
| CRAWL_WORKER            | 8                                                     | Number of worker use for crawl concurrently                                                                                               |
| DOWNLOAD_WORKER         | 1                                                     | Number of workers to download images concurrently image                                                                                   |
| SLEEP                   | 2000                                                  | Sleep time use for each iteration when crawl (in minisecond)                                                                              |
| COVER                   | ''                                                    | (Default: random cover) Cover use for converting EPUB format                                                                              |
| TITLE                   | 'Title'                                               | Title use for converting EPUB format                                                                                                      |
| AUTHOR                  | 'Unknown'                                             | Author use for converting EPUB format                                                                                                     |
| CONVERT_FORMAT          | 'EPUB'                                                | Convert format allow: PDF, EPUB (in-casesensitive)                                                                                        |
