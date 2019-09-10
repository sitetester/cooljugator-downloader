package service

import (
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

type Downloader struct{}

func (downloader Downloader) Download(url string, fileName string, wg *sync.WaitGroup) error {
	defer wg.Done()
	// Create the file
	out, err := os.Create("./files/" + fileName)
	if err != nil {
		return err
	}
	defer out.Close()

	timeout := time.Duration(300 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	// Get the data
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		panic(err)
	}

	return nil
}
