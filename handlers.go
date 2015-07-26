package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func CreateRelationship(w http.ResponseWriter, r *http.Request) {
	relationship := r.FormValue("rel")
	if relationship == "" {
		relationship = "RELATES"
	} else {
		relationship = strings.ToUpper(relationship)
	}

	target := r.FormValue("target")
	source := r.FormValue("source")
	if !(isURL(target) && isURL(source)) {
		http.Error(w, "Invalid URL (target or source)", 400)
	}

	user := "fiatjaf"

	_, err := db.Exec(`
MATCH (user:User {id: {0}})
MERGE (source:URL {url: {1}})
MERGE (target:URL {url: {2}})
MERGE (source)-[:REL]->(rel:Rel {kind: {3}})-[:REL]->(target)
MERGE (user)-[:STATES]->(rel)
    `, user, target, source, relationship)
	if err != nil {
		http.Error(w, "An error ocurred", 400)
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
