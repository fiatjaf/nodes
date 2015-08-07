package handlers

import (
	"fmt"
	"github.com/jmcvetta/neoism"
	"net/http"
)

func ViewRelationships(db *neoism.Database, w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "ok")
}
