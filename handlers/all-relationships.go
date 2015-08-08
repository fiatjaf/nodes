package handlers

import (
	"github.com/jmcvetta/neoism"
	"log"
	"net/http"

	"nodes/helpers"
)

func AllRelationships(db *neoism.Database, w http.ResponseWriter, r *http.Request) {
	// get nodes from database
	res := []struct {
		A neoism.Node
		R neoism.Node
		B neoism.Node
	}{}
	cq := neoism.CypherQuery{
		Statement: `
MATCH (a:Node)-[r:RELATIONSHIP]->(b:Node)
RETURN a AS A, r AS R, b AS B
        `,
		Result: &res,
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
