package handlers

import (
	"github.com/jmcvetta/neoism"
	"log"
	"net/http"
	"net/url"

	"nodes/helpers"
)

func ViewRelationships(db *neoism.Database, w http.ResponseWriter, r *http.Request) {
	log.Print(r.URL.Query().Get("url"))

	// get nodes from database concerning the url requested
	rawurl, err := url.QueryUnescape(r.URL.Query().Get("url"))
	if err != nil {
		http.Error(w, "URL should be escaped.", 400)
		return
	}

	u, err := helpers.ParseURL(rawurl)
	if err != nil {
		http.Error(w, "URL is invalid.", 400)
		return
	}
	stdurl := helpers.GetStandardizedURL(u)

	res := []struct {
		A    string
		B    string
		REL  string
		URLS []string
	}{}
	cq := neoism.CypherQuery{
		Statement: `
MATCH (u:URL)<-[:INSTANCE]-(a) WHERE u.stdurl = {url} OR u.url = {url}
MATCH path=(a)-[r:RELATIONSHIP*1..10]-(b)
WITH nodes(path) AS nodes
UNWIND nodes AS n
WITH DISTINCT n AS n
MATCH (u)<-[:INSTANCE]-(n)
RETURN
  n.name AS a,
  extract(url IN collect(DISTINCT u) | url.stdurl) AS urls,
  '' AS b,
  '' AS rel

UNION ALL

MATCH (u:URL)<-[:INSTANCE]-(a) WHERE u.stdurl = {url} OR u.url = {url}
MATCH path=(a)-[r:RELATIONSHIP*1..10]-(b)
WITH relationships(path) AS rels
UNWIND rels AS r
WITH DISTINCT r AS r
RETURN
  startnode(r).name AS a,
  endnode(r).name AS b,
  r.kind AS rel,
  [] AS urls
        `,
		Parameters: neoism.Props{"url": stdurl},
		Result:     &res,
	}
	err = db.Cypher(&cq)
	if err != nil {
		log.Print(err)
		http.Error(w, "An error ocurred", 400)
		return
	}

	// make dot string
	s := helpers.GenerateDotString(res, r.URL.Query())
	log.Print(s)

	// generate svg graph
	buffer := helpers.RenderGraph(s)

	// write the response as image
	w.Header().Set("Content-Type", "image/svg+xml")
	if _, err := w.Write(buffer.Bytes()); err != nil {
		log.Print("unable to write image. ", err)
	}
}
