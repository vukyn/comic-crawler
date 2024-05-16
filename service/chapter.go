package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
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

func LoadChapters() ([]Chapter, error) {
	comicId := os.Getenv("COMIC_ID")
	domain := os.Getenv("DOMAIN")
	query := os.Getenv("COMIC_QUERY")
	url := "https://www." + domain + query + comicId

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

	return chapterResponse.Chapters, nil
}
