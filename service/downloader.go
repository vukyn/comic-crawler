package service

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// DownloadImage downloads an image from the given URL and saves it to the specified path
func DownloadImage(workerId int, url string, filepath string) error {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()
	// Get the image data
	t := time.Now()
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	fmt.Printf("(Worker %d) Downloading image: %v - Took (%.2fs)\n", workerId, url, time.Since(t).Seconds())

	// Check for successful response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download image: %s", url)
	}

	// Copy data from response to file
	_, err = io.Copy(out, resp.Body)
	return err
}
