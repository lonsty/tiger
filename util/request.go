package util

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/PuerkitoBio/goquery"
)

// Request HTTP request 请求
type Request struct {
	Headers map[string]string
	Cookies map[string]string
}

// Get http GET 请求
func (r *Request) Get(url string) (*http.Response, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Get %s error: %v", url, err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Get %s status code: %v", url, resp.StatusCode)
	}
	return resp, nil
}

// GetToFile HTTP GET 请求并下载文件
func (r *Request) GetToFile(url, filename string) error {
	resp, err := r.Get(url)
	if err != nil {
		return fmt.Errorf("GetToFile %s error: %v", url, err)
	}
	defer resp.Body.Close()

	out, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("GetToFile %s create file %s error: %v", url, filename, err)
	}
	defer out.Close()

	n, err := io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("GetToFile %s copy response error: %v", url, err)
	}
	log.Printf("Saved %s to %s", ByteCountIEC(n), filename)
	return nil
}

// GetToDocument HTTP GET 请求并加载到 HTML DOM
func (r *Request) GetToDocument(url string) (*goquery.Document, error) {
	resp, err := r.Get(url)
	if err != nil {
		return nil, fmt.Errorf("GetToDocument %s error: %v", url, err)
	}
	defer resp.Body.Close()

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("GetToDocument %s load html document error: %v", url, err)
	}
	return doc, nil
}
