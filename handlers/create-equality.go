package handlers

import (
	"github.com/jmcvetta/neoism"
	"github.com/kr/pretty"
	"log"
	"net/http"
	"time"

	"nodes/helpers"
)

func CreateEquality(db *neoism.Database, w http.ResponseWriter, r *http.Request) {
	source, err := helpers.ParseURL(r.FormValue("source"))
	if err != nil {
		http.Error(w, "source is invalid URL: "+r.FormValue("source"), 400)
		return
	}
	target, err := helpers.ParseURL(r.FormValue("target"))
	if err != nil {
		http.Error(w, "target is invalid URL: "+r.FormValue("target"), 400)
		return
	}
	sourceTitle, err := helpers.GetTitle(source)
	if err != nil {
		http.Error(w, "Couldn't fetch title for "+source.String(), 400)
		return
	}
	targetTitle, err := helpers.GetTitle(target)
	if err != nil {
		http.Error(w, "Couldn't fetch title for "+target.String(), 400)
		return
	}
	if helpers.EqualURLs(source, target) {
		http.Error(w, "urls are equal.", 400)
		return
	}

	// standardize urls
	stdsource := helpers.GetStandardizedURL(source)
	stdtarget := helpers.GetStandardizedURL(target)

	// get user
	user := "fiatjaf"
	// user.GetKarma()

	now := time.Now().UTC().Format("20060102150405")

	var queries []*neoism.CypherQuery
	res := []struct {
		SN            int
		TN            int
		SU            int
		TU            int
		RELATIONSHIPS []int
	}{}
	cq := neoism.CypherQuery{
		Statement: `
MERGE (su:URL {stdurl: {stdsource}}) ON CREATE SET su.title = {sourceTitle}
MERGE (tu:URL {stdurl: {stdtarget}}) ON CREATE SET tu.title = {targetTitle}

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
			"stdsource":   stdsource,
			"stdtarget":   stdtarget,
			"sourceTitle": sourceTitle,
			"targetTitle": targetTitle,
		},
		Result: &res,
	}
	queries = append(queries, &cq)

	txn, err := db.Begin(queries)
	if err != nil {
		pretty.Log(err)
		http.Error(w, err.Error(), 500)
		txn.Rollback()
		return
	}

	row := res[0]

	// do our checks to decide if we are going to create Nodes, mix them or what
	queries = make([]*neoism.CypherQuery, 0)
	if row.SN == row.TN && row.SN != 0 {
		// they exist and are the same, so we do nothing
	} else if row.SN != 0 && row.TN != 0 {
		// both exists, transfer everything from the source to target and delete source
		for _, r := range row.RELATIONSHIPS {
			log.Print(row.SN, row.TN, r)
			queries = append(queries, &neoism.CypherQuery{
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
		queries = append(queries, &neoism.CypherQuery{
			Statement: `
MATCH (sn) WHERE id(sn) = {sn}
MATCH (tn) WHERE id(tn) = {tn}
MATCH (su) WHERE id(su) = {su}

MATCH (sn)-[oldinstance:INSTANCE]->(su)
MERGE (tn)-[newinstance:INSTANCE]->(su)
ON CREATE SET newinstance = oldinstance
ON CREATE SET newinstance.user = {user}

WITH oldinstance, sn
MATCH (sn)-[srels:RELATIONSHIP]-()
DELETE oldinstance, sn, srels
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
		queries = append(queries, &neoism.CypherQuery{
			Statement: `
MATCH (su) WHERE id(su) = {su}
MATCH (tu) WHERE id(tu) = {tu}

CREATE (n:Node {created: {now}, name: tu.title})
MERGE (n)-[r1:INSTANCE {user: {user}}]->(su) ON CREATE SET r1.created = {now}
MERGE (n)-[r2:INSTANCE {user: {user}}]->(tu) ON CREATE SET r2.created = {now}
            `,
			Parameters: neoism.Props{
				"su":   row.SU,
				"tu":   row.TU,
				"user": user,
				"now":  now,
			},
		})
		pretty.Log(queries)
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
		queries = append(queries, &neoism.CypherQuery{
			Statement: `
MATCH (floating) WHERE id(floating) = {floating}
MATCH (appendTo) WHERE id(appendTo) = {appendTo}

MERGE (appendTo)-[r:INSTANCE {user: {user}}]->(floating) ON CREATE SET r.created = {now}
            `,
			Parameters: neoism.Props{
				"floating": floating,
				"appendTo": appendTo,
				"user":     user,
				"now":      now,
			},
		})
	}

	err = txn.Query(queries)
	if err != nil {
		pretty.Log(err)
		pretty.Log(queries)
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

	w.WriteHeader(http.StatusOK)
}
