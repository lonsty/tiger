/*
Copyright © 2022 lonsty <lonsty@sina.com>

*/
package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/lonsty/tiger/util"
	"github.com/spf13/cobra"
)

var directory string

// imageCmd represents the image command
var imageCmd = &cobra.Command{
	Use:   "image",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: DownloadImages,
}

func init() {
	rootCmd.AddCommand(imageCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// imageCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	imageCmd.PersistentFlags().StringVarP(&directory, "directory", "d", ".",
		"Destination to save the images")
}

// DownloadImages 批量下载图片
func DownloadImages(cmd *cobra.Command, args []string) {
	var wg sync.WaitGroup
	for idx, postURL := range args {
		log.Printf("idx: %d, postURL: %s", idx, postURL)
		postURL := postURL
		go func() {
			wg.Add(1)
			defer wg.Done()
			log.Printf("postURL: %s", postURL)
			imagePost, err := DownloadImagesFromPostURL(postURL, directory)
			if err != nil {
				log.Printf("DownloadImages error: %v", err)
				return
			}
			log.Printf("%s %s succeed: %d, failed: %d",
				imagePost.Category, imagePost.Title, len(imagePost.DownloadSucceed), len(imagePost.DownloadFailed))
		}()
	}
	wg.Wait()
}

// DownloadImagesFromPostURL 从帖子批量下载图片
func DownloadImagesFromPostURL(postURL, directory string) (*ImagePost, error) {
	req := new(util.Request)
	doc, err := req.GetToDocument(postURL)
	if err != nil {
		return nil, fmt.Errorf("DownloadImagesFromPostURL error: %v", err)
	}

	imagePost := ParseDocumentToImagePost(doc)
	dir := filepath.Join(directory, imagePost.Category, imagePost.Title)
	log.Printf("dir: %s", dir)
	_ = os.MkdirAll(dir, os.ModePerm)

	var wg sync.WaitGroup
	for idx, imageURL := range imagePost.ImageURLs {
		idx := idx
		imageURL := imageURL
		go func() {
			wg.Add(1)
			defer wg.Done()
			log.Printf("idx: %d, imageURL: %s", idx, imageURL)
			parts := strings.Split(imageURL, "/")
			filename := filepath.Join(dir, fmt.Sprintf("[%02d]%s", idx+1, parts[len(parts)-1]))
			if _, err := os.Stat(filename); !errors.Is(err, os.ErrNotExist) {
				// 文件存在
				log.Printf("Already downloaded: %s", filename)
				return
			}
			// download to file
			if err := req.GetToFile(imageURL, filename); err != nil {
				imagePost.DownloadFailed = append(imagePost.DownloadFailed, imageURL)
				log.Printf("DownloadImagesFromPostURL idx: %d, imageURL: %s error: %v", idx, imageURL, err)
			} else {
				imagePost.DownloadSucceed = append(imagePost.DownloadSucceed, imageURL)
			}
		}()
	}
	wg.Wait()
	return imagePost, nil
}

// ImagePost 图片帖子
type ImagePost struct {
	Category        string   // 帖子分类
	Title           string   // 帖子标题
	PostAt          string   // 发布时间
	ImageURLs       []string // 帖子图片链接
	DownloadSucceed []string // 下载成功的图片链接
	DownloadFailed  []string // 下载的失败的图片链接
}

func ParseDocumentToImagePost(doc *goquery.Document) *ImagePost {
	mainTag := doc.Find(".main").First()
	title := mainTag.Find("h1").First().Nodes[0].FirstChild.Data
	categoryAndDate := mainTag.Find("h2").First().Nodes[0].FirstChild.Data
	fields := strings.Fields(categoryAndDate)

	// Find the review items
	var imageURLs []string
	doc.Find(".pic img").Each(func(i int, s *goquery.Selection) {
		imageURL := s.Nodes[0].Attr[0].Val
		imageURLs = append(imageURLs, imageURL)
	})

	return &ImagePost{
		Category:        strings.TrimSpace(fields[0]),
		Title:           strings.TrimSpace(title),
		PostAt:          strings.TrimSpace(fields[1]),
		ImageURLs:       imageURLs,
		DownloadSucceed: []string{},
		DownloadFailed:  []string{},
	}
}
