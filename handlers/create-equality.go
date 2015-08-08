package handlers

import (
	"github.com/jmcvetta/neoism"
	"github.com/kr/pretty"
	"net/http"
	"time"

	"nodes/helpers"
)

func CreateEquality(db *neoism.Database, w http.ResponseWriter, r *http.Request) {
	source := r.FormValue("source")
	if !helpers.IsURL(source) {
		http.Error(w, "source is invalid URL: "+source, 400)
		return
	}
	target := r.FormValue("target")
	if !helpers.IsURL(target) {
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
	// user.GetKarma()

	now := time.Now().UTC().Format("20060102150405")

	var qs []*neoism.CypherQuery
	res := []struct {
		SN            int
		TN            int
		SU            int
		TU            int
		RELATIONSHIPS []int
	}{}
	cq := neoism.CypherQuery{
		Statement: `
MERGE (su:URL {url: {source}}) ON CREATE SET su.title = {sourceTitle}
MERGE (tu:URL {url: {target}}) ON CREATE SET tu.title = {targetTitle}

WITH su, tu
OPTIONAL MATCH (su)<-[:INSTANCE]-(sn:Node)
OPTIONAL MATCH (tu)<-[:INSTANCE]-(tn:Node)
OPTIONAL MATCH (sn)-[r:RELATIONSHIP]-()

WITH sn, tn, su, tu, collect(r) AS rels
RETURN id(sn) AS sn,
       id(tn) AS tn,
       id(su) AS su,
       id(tu) AS tu,
       extract(rel IN rels | id(rel)) AS relationships
        `,
		Parameters: neoism.Props{
			"source":      source,
			"target":      target,
			"sourceTitle": sourceTitle,
			"targetTitle": targetTitle,
		},
		Result: &res,
	}
	qs = append(qs, &cq)

	txn, err := db.Begin(qs)
	if err != nil {
		pretty.Log(err)
		http.Error(w, err.Error(), 500)
		txn.Rollback()
		return
	}

	row := res[0]

	// do our checks to decide if we are going to create Nodes, mix them or what
	qs = make([]*neoism.CypherQuery, 0)
	if row.SN != 0 && row.TN != 0 {
		// both exists, transfer everything from the source to target and delete source
		for _, r := range row.RELATIONSHIPS {
			pretty.Log(row.SN, row.TN, r)
			qs = append(qs, &neoism.CypherQuery{
				Statement: `
MATCH (sn) WHERE id(sn) = {sn}
MATCH (tn) WHERE id(tn) = {tn}
MATCH (a)-[r]->(b) WHERE id(r) = {r}

FOREACH (x IN CASE WHEN a = sn THEN [1] ELSE [] END |
  MERGE (tn)-[newrel:RELATIONSHIP {user: r.user}]->(b)
  ON CREATE SET newrel.created = r.created
  SET newrel.kind = r.kind
)

FOREACH (x IN CASE WHEN b = sn THEN [1] ELSE [] END |
  MERGE (a)<-[newrel:RELATIONSHIP {user: r.user}]-(tn)
  ON CREATE SET newrel.created = r.created
  SET newrel.kind = r.kind
)
                `,
				Parameters: neoism.Props{
					"sn": row.SN,
					"tn": row.TN,
					"r":  r,
				},
			})
		}
		qs = append(qs, &neoism.CypherQuery{
			Statement: `
MATCH (sn) WHERE id(sn) = {sn}
MATCH (tn) WHERE id(tn) = {tn}
MATCH (su) WHERE id(su) = {su}

MATCH (sn)-[oldinstance:INSTANCE]->(su)
MERGE (tn)-[newinstance:INSTANCE]->(su)
ON CREATE SET newinstance = oldinstance
ON CREATE SET newinstance.user = {user}

WITH oldinstance, sn
DELETE oldinstance, sn
            `,
			Parameters: neoism.Props{
				"sn":   row.SN,
				"tn":   row.TN,
				"su":   row.SU,
				"user": user,
			},
		})
	} else if row.SN == 0 && row.TN == 0 {
		// none exist, create one for both
		qs = append(qs, &neoism.CypherQuery{
			Statement: `
MATCH (su) WHERE id(su) = {su}
MATCH (tu) WHERE id(tu) = {tu}

CREATE (n:Node {created: {now}, name: tu.title})
CREATE (n)-[:INSTANCE {user: {user}, created: {now}}]->(su)
CREATE (n)-[:INSTANCE {user: {user}, created: {now}}]->(tu)
            `,
			Parameters: neoism.Props{
				"su":   row.SU,
				"tu":   row.TU,
				"user": user,
				"now":  now,
			},
		})
	} else {
		var floating int
		var appendTo int
		if row.SN != 0 {
			// only SN exist, append TU to SN
			floating = row.TU
			appendTo = row.SN

		} else if row.TN != 0 {
			// only TN exist, append SU to TN
			floating = row.SU
			appendTo = row.TN
		}
		qs = append(qs, &neoism.CypherQuery{
			Statement: `
MATCH (floating) WHERE id(floating) = {floating}
MATCH (appendTo) WHERE id(appendTo) = {appendTo}

CREATE (appendTo)-[:INSTANCE {user: {user}, created: {now}}]->(floating)
            `,
			Parameters: neoism.Props{
				"floating": floating,
				"appendTo": appendTo,
				"user":     user,
				"now":      now,
			},
		})
	}

	err = txn.Query(qs)
	if err != nil {
		pretty.Log(err)
		http.Error(w, err.Error(), 500)
		txn.Rollback()
		return
	}

	err = txn.Commit()
	if err != nil {
		pretty.Log(err)
		http.Error(w, err.Error(), 500)
		txn.Rollback()
		return
	}

	http.Redirect(w, r, "/rels.svg", 302)
}
