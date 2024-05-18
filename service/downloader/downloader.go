package downloader

import (
	"bytes"
	"comic-crawler/env"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/vukyn/kuery/log"
)

// DownloadImg downloads an image from the given URL and saves it to the specified path
func DownloadImg(workerId int, url, domain, filepath string) error {
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
	log.Infof("(Worker %d) Downloaded image: %v - Took (%.2fs)", workerId, url, time.Since(t).Seconds())

	// Copy data from response to file
	_, err = io.Copy(out, bytes.NewReader(res))
	return err
}

func getReferer(domain string) string {
	var referer = map[string]string{
		env.NettruyenDomain: env.NettruyenReferer,
		env.QqtruyenDomain:  env.QqtruyenReferer,
	}
	return referer[domain]
}

func makeGet(url string, header map[string]string) ([]byte, error) {
	ctx, cancel := context.WithDeadline(context.TODO(), time.Now().Add(2*time.Second))
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		log.Errorf("Error creating request: %v", err)
		return nil, err
	}
	for key, value := range header {
		req.Header.Add(key, value)
	}

	data, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Errorf("Error making request: %v", err)
		return nil, err
	}
	defer data.Body.Close()

	// Check for successful response
	if data.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download image: %s", url)
	}

	res, err := io.ReadAll(data.Body)
	if err != nil {
		log.Errorf("Error reading response: %v", err)
		return nil, err
	}

	return res, nil
}
