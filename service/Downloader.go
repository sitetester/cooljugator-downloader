package service

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"sync"
	"time"
)

type Downloader struct{}

func (downloader Downloader) Download(dir string, url string, fileName string, wg *sync.WaitGroup) {
	defer wg.Done()

	fmt.Println("Downloading...", url)

	// Create New http Transport
	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // disable verify
		},
	}

	// Create Http Client
	timeout := time.Duration(300 * time.Second)
	client := &http.Client{
		Transport: transCfg,
		Timeout:   timeout,
	}

	// Request
	resp, err := client.Get(url)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	newPath := filepath.Join(dir, fileName)

	ioutil.WriteFile(newPath, body, 0644)
}
