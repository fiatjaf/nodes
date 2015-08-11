package helpers

import (
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/url"
	"strings"
)

func ParseURL(potentialURL string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(potentialURL))
	u.Fragment = ""
	if err != nil {
		return "", err
	}
	if u.Scheme == "" {
		return "", err
	}
	if u.Host == "" {
		return "", err
	}
	return u.String(), nil
}

func EqualURLs(source string, target string) bool {
	if source == target {
		return true
	}
	return false
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

	return strings.TrimSpace(doc.Find("title").First().Text()), nil
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
