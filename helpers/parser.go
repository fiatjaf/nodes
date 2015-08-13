package helpers

import (
	"encoding/json"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func GetTitle(u *url.URL) (string, error) {
	title := getTitleFromReadability(u)
	if title != "" {
		return title, nil
	}

	doc, err := goquery.NewDocument(u.String())
	if err != nil {
		return getLastPathPart(u)
	}
	title = strings.TrimSpace(doc.Find("title").First().Text())
	if title == "" {
		return getLastPathPart(u)
	}
	return title, nil
}

func getTitleFromReadability(u *url.URL) string {
	// build querystring
	qs := url.Values{}
	qs.Set("url", u.String())
	qs.Set("token", os.Getenv("READABILITY_PARSER_KEY"))

	// http get
	resp, err := http.Get("https://readability.com/api/content/v1/parser?" + qs.Encode())
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	// get the json results (only the title)
	var data struct {
		Title string `json: "title"`
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return ""
	}
	return data.Title
}

func getLastPathPart(u *url.URL) (string, error) {
	path := strings.Split(u.Path, "/")
	path = Filter(path, func(s string) bool {
		if s == "" {
			return false
		}
		return true
	})
	return path[len(path)-1], nil
}
