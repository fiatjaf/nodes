package handlers

import (
	"bytes"
	"fmt"
	"github.com/awalterschulze/gographviz"
	"github.com/jmcvetta/neoism"
	"log"
	"net/http"
	"os/exec"
	"strings"
)

func ViewRelationships(db *neoism.Database, w http.ResponseWriter, r *http.Request) {
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
	g := gographviz.NewGraph()
	g.SetName("nodes")
	g.SetDir(true)

	for _, row := range res {
		aname := fmt.Sprintf("\"%s\"", row.A.Data["name"].(string))
		rkind := row.R.Data["kind"].(string)
		bname := fmt.Sprintf("\"%s\"", row.B.Data["name"].(string))
		g.AddNode("nodes", aname, map[string]string{
			"label": aname,
		})
		g.AddNode("nodes", bname, map[string]string{
			"label": bname,
		})
		g.AddEdge(aname, bname, true, map[string]string{
			"label": rkind,
		})
	}

	s := g.String()
	log.Print(s)

	// generate svg graph
	var out bytes.Buffer
	cmd := exec.Command("dot", "-Tsvg")
	cmd.Stdin = strings.NewReader(s)
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		log.Print(err.Error())
	}

	// write the response as image
	w.Header().Set("Content-Type", "image/svg+xml")
	if _, err := w.Write(out.Bytes()); err != nil {
		log.Print("unable to write image. ", err)
	}
}
