package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/PuerkitoBio/goquery"
	"regexp"
)

func main() {
	url := os.Args[1]
	for url != "#" {
		fmt.Printf("Scraping %s...", url)
		imgURL, nextURL, err := scrapeImgAndNext(url)
		if err != nil {
			fmt.Println(err)
			fmt.Println("Retry...")
			continue
		}
		err = download(imgURL)
		if err != nil {
			fmt.Println(err)
			fmt.Println("Retry...")
			continue
		}
		url = nextURL
		fmt.Println("done")
	}
}

var rePage = regexp.MustCompile(`/page.php\?.*`)

func scrapeImgAndNext(rawurl string) (img string, next string, err error) {
	doc, err := goquery.NewDocument(rawurl)
	if err != nil {
		return "", "", err
	}
	aNode := doc.Find("div#maincontent a")
	if aNode.Length() == 0 {
		return "", "", fmt.Errorf(`Can't find "div#maincontent a" node`)
	}
	nextURL, ok := aNode.Attr("href")
	if !ok {
		return "", "", fmt.Errorf(`Can't find "href" attr`)
	}
	if nextURL != "#" {
		nextURL = rePage.ReplaceAllString(rawurl, "/"+nextURL)
	}
	imgURL, ok := aNode.Find("img").Attr("src")
	if !ok {
		return "", "", fmt.Errorf("Can't find img url")
	}
	return imgURL, nextURL, nil
}

func download(rawurl string) error {
	filename, err := fileNameOf(rawurl)
	if err != nil {
		return err
	}
	resp, err := http.Get(rawurl)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

var reInPath = regexp.MustCompile("[^/]+$")

func fileNameOf(rawurl string) (string, error) {
	url, err := url.Parse(rawurl)
	if err != nil {
		return "", err
	}
	file := reInPath.FindString(url.Path)
	if file == "" {
		return "", fmt.Errorf("Filename not found: %s", rawurl)
	}
	return file, nil
}
