package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func CreateRelationship(w http.ResponseWriter, r *http.Request) {
	relationship := r.FormValue("rel")
	if relationship == "" {
		relationship = "relates"
	} else {
		relationship = strings.ToLower(relationship)
	}

	target := r.FormValue("target")
	source := r.FormValue("source")
	if !(isURL(target) && isURL(source)) {
		http.Error(w, "Invalid URL (target or source)", 400)
		return
	}

	sourceTitle, err := getTitle(source)
	if err != nil {
		http.Error(w, "Couldn't fetch title for "+source, 400)
		return
	}
	targetTitle, err := getTitle(target)
	if err != nil {
		http.Error(w, "Couldn't fetch title for "+target, 400)
		return
	}

	user := "fiatjaf"

	// sn, su, tn, tu = source node, source url, target node, target url
	_, err = db.Exec(`
MERGE (su:URL {url: {0}})
MERGE (tu:URL {url: {1}})
SET su.title = {3}
SET tu.title = {4}

CREATE UNIQUE (su)<-[:INSTANCE]-(sn:Node)
CREATE UNIQUE (tu)<-[:INSTANCE]-(tn:Node)
SET sn.name = CASE WHEN sn.name IS NOT NULL THEN sn.name ELSE su.title END
SET tn.name = CASE WHEN tn.name IS NOT NULL THEN tn.name ELSE tu.title END

MERGE (sn)-[rel:RELATES {user: {5}}]->(tn)
SET rel.kind = {2}
    `, source, target, relationship, sourceTitle, targetTitle, user)
	if err != nil {
		log.Print(err)
		http.Error(w, "An error ocurred", 400)
		return
	}

	http.Redirect(w, r, r.URL.String(), 302)
}

func ViewRelationships(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "ok")
}

func isURL(potentialURL string) bool {
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

func getTitle(URL string) (string, error) {
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
