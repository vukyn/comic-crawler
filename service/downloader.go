package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// DownloadImage downloads an image from the given URL and saves it to the specified path
func DownloadImage(workerId int, url, domain, filepath string) error {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the image data
	header := make(map[string]string)
	referer := getReferer(domain)
	if referer != "" {
		header["Referer"] = referer
	}
	t := time.Now()
	res, err := makeGet(url, header)
	if err != nil {
		return err
	}
	fmt.Printf("(Worker %d) Downloading image: %v - Took (%.2fs)\n", workerId, url, time.Since(t).Seconds())

	// Copy data from response to file
	_, err = io.Copy(out, bytes.NewReader(res))
	return err
}

func getReferer(domain string) string {
	var referer = map[string]string{
		os.Getenv("NETTRUYEN_DOMAIN"): os.Getenv("NETTRUYEN_REFERER"),
		os.Getenv("QQTRUYEN_DOMAIN"):  os.Getenv("QQTRUYEN_REFERER"),
	}
	return referer[domain]
}

func makeGet(url string, header map[string]string) ([]byte, error) {
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return nil, err
	}
	for key, value := range header {
		req.Header.Add(key, value)
	}

	data, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return nil, err
	}
	defer data.Body.Close()

	// Check for successful response
	if data.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download image: %s", url)
	}

	res, err := io.ReadAll(data.Body)
	if err != nil {
		fmt.Println("Error reading data:", err)
		return nil, err
	}

	return res, nil
}
