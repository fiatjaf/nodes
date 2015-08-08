package handlers

import (
	"github.com/jmcvetta/neoism"
	"log"
	"net/http"

	"nodes/helpers"
)

func ViewRelationships(db *neoism.Database, w http.ResponseWriter, r *http.Request) {
	// get nodes from database concerning the url requested
	url := r.URL.Query().Get("url")

	res := []struct {
		A neoism.Node
		R neoism.Node
		B neoism.Node
	}{}
	cq := neoism.CypherQuery{
		Statement: `
MATCH (u {url: {url}})
MATCH (u)<-[:INSTANCE]-(a:Node)
MATCH path=(a)-[r:RELATIONSHIP*1..10]-(b:Node)
WITH relationships(path) as rels
UNWIND rels AS r
RETURN DISTINCT startnode(r) AS A, r AS R, endnode(r) AS B
        `,
		Parameters: neoism.Props{"url": url},
		Result:     &res,
	}
	err := db.Cypher(&cq)
	if err != nil {
		log.Print(err)
		http.Error(w, "An error ocurred", 400)
		return
	}

	// make dot string
	s := helpers.GenerateDotString(res)
	log.Print(s)

	// generate svg graph
	buffer := helpers.RenderGraph(s)

	// write the response as image
	w.Header().Set("Content-Type", "image/svg+xml")
	if _, err := w.Write(buffer.Bytes()); err != nil {
		log.Print("unable to write image. ", err)
	}
}
