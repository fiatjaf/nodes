package helpers

import (
	"errors"
	"github.com/PuerkitoBio/goquery"
	"net/url"
	"strings"
)

// {host: true}
var uselessPath map[string]bool

// {host+path: [parametername]}
var importantQueryParam map[string][]string

func init() {
	uselessPath = make(map[string]bool)
	uselessPath["books.google.com"] = true
	uselessPath["books.google.com.br"] = true

	importantQueryParam = make(map[string][]string)
	importantQueryParam["news.ycombinator.com/user"] = []string{"id"}
	importantQueryParam["news.ycombinator.com/threads"] = []string{"id"}
	importantQueryParam["news.ycombinator.com/submitted"] = []string{"id"}
	importantQueryParam["news.ycombinator.com/saved"] = []string{"id", "comments"}
	importantQueryParam["youtube.com/watch"] = []string{"v"}
	importantQueryParam["youtube.com/playlist"] = []string{"list"}
	importantQueryParam["books.google.com"] = []string{"id"}
	importantQueryParam["books.google.com.br"] = []string{"id"}
	importantQueryParam["drive.google.com/folderview"] = []string{"id"}
}

func ParseURL(potentialURL string) (*url.URL, error) {
	u, err := url.Parse(strings.TrimSpace(potentialURL))
	if err != nil {
		return u, err
	}
	if u.Scheme == "" {
		u.Scheme = "http"
	}
	if u.Host == "" {
		return u, errors.New("URL is not absolute.")
	}
	u.Fragment = ""
	u.Host = strings.ToLower(u.Host)
	u.Scheme = strings.ToLower(u.Scheme)
	return u, nil
}

func EqualURLs(source *url.URL, target *url.URL) bool {
	if source.String() == target.String() {
		return true
	}
	return false
}

func GetStandardizedURL(u *url.URL) string {
	if strings.HasPrefix(u.Host, "www.") {
		// remove starting www.
		u.Host = u.Host[4:]
	}
	if strings.HasSuffix(u.Path, "/") {
		// remove ending slash
		u.Path = u.Path[:len(u.Path)-1]
	}
	if u.Scheme == "https" {
		// always 'http'
		u.Scheme = "http"
	}
	// remove path if it not significant
	if _, ok := uselessPath[u.Host]; ok {
		u.Path = ""
	}

	// leave only significant querystring parameters
	qs := u.Query()
	hostpath := u.Host + u.Path
	var stdrawqs string
	if params, ok := importantQueryParam[hostpath]; ok {
		// this hostpath has significant querystring parameters
		// build the new standardized querystring
		stdqs := url.Values{}
		for _, param := range params {
			stdqs[param] = qs[param]
		}
		stdrawqs = stdqs.Encode()
	} else {
		// this hostpath has no significant querystring parameters
		// the new standardized querystring is an empty string
		stdrawqs = ""
	}
	u.RawQuery = stdrawqs

	return u.String()
}

func GetTitle(u *url.URL) (string, error) {
	doc, err := goquery.NewDocument(u.String())
	if err != nil {
		return getLastPathPart(u)
	}
	title := strings.TrimSpace(doc.Find("title").First().Text())
	if title == "" {
		return getLastPathPart(u)
	}
	return title, nil
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

func Filter(s []string, fn func(string) bool) []string {
	var p []string
	for _, v := range s {
		if fn(v) {
			p = append(p, v)
		}
	}
	return p
}
