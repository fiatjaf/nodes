package handlers

import (
	"github.com/jmcvetta/neoism"
	"log"
	"net/http"
	"strings"
	"time"

	"nodes/helpers"
)

func CreateRelationship(db *neoism.Database, w http.ResponseWriter, r *http.Request) {
	relationship := r.FormValue("rel")
	if relationship == "" {
		relationship = "relates"
	} else {
		relationship = strings.ToLower(relationship)
	}

	source, err := helpers.ParseURL(r.FormValue("source"))
	if err != nil {
		http.Error(w, "source is invalid URL: "+source, 400)
		return
	}
	target, err := helpers.ParseURL(r.FormValue("target"))
	if err != nil {
		http.Error(w, "target is invalid URL: "+target, 400)
		return
	}

	sourceTitle, err := helpers.GetTitle(source)
	if err != nil {
		http.Error(w, "Couldn't fetch title for "+source, 400)
		return
	}
	targetTitle, err := helpers.GetTitle(target)
	if err != nil {
		http.Error(w, "Couldn't fetch title for "+target, 400)
		return
	}

	user := "fiatjaf"

	cq := neoism.CypherQuery{
		// sn, su, tn, tu = source node, source url, target node, target url
		Statement: `
MERGE (su:URL {url: {source}})
MERGE (tu:URL {url: {target}})
SET su.title = {sourceTitle}
SET tu.title = {targetTitle}

CREATE UNIQUE (su)<-[:INSTANCE]-(sn:Node)
CREATE UNIQUE (tu)<-[:INSTANCE]-(tn:Node)
SET sn.name = CASE WHEN sn.name IS NOT NULL THEN sn.name ELSE su.title END
SET tn.name = CASE WHEN tn.name IS NOT NULL THEN tn.name ELSE tu.title END

MERGE (sn)-[rel:RELATIONSHIP {user: {user}}]->(tn)
ON CREATE SET rel.created = {now}
SET rel.kind = {relationshipKind}
        `,
		Parameters: neoism.Props{
			"source":           source,
			"target":           target,
			"sourceTitle":      sourceTitle,
			"targetTitle":      targetTitle,
			"user":             user,
			"relationshipKind": relationship,
			"now":              time.Now().UTC().Format("20060102150405"),
		},
	}
	err = db.Cypher(&cq)
	if err != nil {
		log.Print(err)
		http.Error(w, "An error ocurred", 400)
		return
	}

	http.Redirect(w, r, "/cluster.svg?url="+source, 302)
}