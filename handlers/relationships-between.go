package handlers

import (
	"github.com/jmcvetta/neoism"
	"log"
	"net/http"
	"net/url"

	"nodes/helpers"
)

func RelationshipsBetween(db *neoism.Database, w http.ResponseWriter, r *http.Request) {
	// check existence of variables
	var rawsource string
	var rawtarget string
	var err error
	qs := r.URL.Query()
	rawsource, err = url.QueryUnescape(qs.Get("source"))
	if err != nil {
		http.Error(w, "source url should be escaped.", 400)
		return
	}
	source, err := helpers.ParseURL(rawsource)
	if err != nil {
		http.Error(w, "source is invalid URL: "+rawsource, 400)
		return
	}
	rawtarget, err = url.QueryUnescape(qs.Get("target"))
	if err != nil {
		http.Error(w, "target url should be escaped.", 400)
		return
	}
	target, err := helpers.ParseURL(rawtarget)
	if err != nil {
		http.Error(w, "target is invalid URL: "+rawtarget, 400)
		return
	}
	if helpers.EqualURLs(source, target) {
		http.Error(w, "urls are equal.", 400)
		return
	}

	// standardize urls
	stdsource := helpers.GetStandardizedURL(source)
	stdtarget := helpers.GetStandardizedURL(target)

	log.Print(source, " ", stdsource)
	log.Print(target, " ", stdtarget)
}
