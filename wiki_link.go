package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func crawl() {
	var res []struct{} = []string{}
	client := &http.Client{Timeout: time.Second * 10}
	url := "https://en.wikipedia.org/wiki/Go_(programming_language)"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}

	req.Header.Set(
		"User-Agent",
		"CodeItCrawler/1.0 (https://github.com/pratyaksh1001/code-it-golang; pratyaksh@example.com)",
	)

	resp, _ := client.Do(req)
	defer resp.Body.Close()
	doc, _ := goquery.NewDocumentFromReader(resp.Body)
	//fmt.Println(doc.Html())
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, exist := s.Attr("href")
		name := s.Text()
		if !exist {
			fmt.Println("empty link")
		} else {
			res = append(res, href)
		}
	})
	for _, i := range res {
		fmt.Println(i + "\n")
	}
}
