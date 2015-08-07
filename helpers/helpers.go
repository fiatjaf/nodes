package helpers

import (
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/url"
	"strings"
)

func IsURL(potentialURL string) bool {
	u, err := url.Parse(potentialURL)
	if err != nil {
		return false
	}
	if u.Scheme == "" {
		return false
	}
	if u.Host == "" {
		return false
	}
	return true
}

func GetTitle(URL string) (string, error) {
	doc, err := goquery.NewDocument(URL)
	if err != nil {
		u, err := url.Parse(URL)
		if err != nil {
			log.Print("invalid url:", URL)
			return "", err
		}
		path := strings.Split(u.Path, "/")
		path = Filter(path, func(s string) bool {
			if s == "" {
				return false
			}
			return true
		})
		return path[len(path)-1], nil
	}

	return doc.Find("title").First().Text(), nil
}

func Filter(s []string, fn func(string) bool) []string {
	var p []string
	for _, v := range s {
		if fn(v) {
			p = append(p, v)
		}
	}
	return p
}
