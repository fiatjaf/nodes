package helpers

import (
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/awalterschulze/gographviz"
	"github.com/jmcvetta/neoism"
	"log"
	"net/url"
	"os/exec"
	"strings"
)

func ParseURL(potentialURL string) (string, error) {
	u, err := url.Parse(potentialURL)
	u.Fragment = ""
	if err != nil {
		return "", err
	}
	if u.Scheme == "" {
		return "", err
	}
	if u.Host == "" {
		return "", err
	}
	return u.String(), nil
}

func GetTitle(URL string) (string, error) {
	doc, err := goquery.NewDocument(URL)
	if err != nil {
		u, err := url.Parse(URL)
		if err != nil {
			log.Print("invalid url:", URL)
			return "", err
		}
		path := strings.Split(u.Path, "/")
		path = Filter(path, func(s string) bool {
			if s == "" {
				return false
			}
			return true
		})
		return path[len(path)-1], nil
	}

	return strings.TrimSpace(doc.Find("title").First().Text()), nil
}

func Filter(s []string, fn func(string) bool) []string {
	var p []string
	for _, v := range s {
		if fn(v) {
			p = append(p, v)
		}
	}
	return p
}

func RenderGraph(dot string) bytes.Buffer {
	var out bytes.Buffer
	cmd := exec.Command("dot", "-Tsvg")
	cmd.Stdin = strings.NewReader(dot)
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Print(err.Error())
	}
	return out
}

func GenerateDotString(res []struct {
	A neoism.Node
	R neoism.Node
	B neoism.Node
}) string {
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
			"dir":   RelationshipDir(rkind),
		})
	}

	return g.String()
}

func RelationshipDir(kind string) string {
	undirectedKinds := map[string]bool{
		"related":      true,
		"similar":      true,
		"same_purpose": true,
	}
	if _, ok := undirectedKinds[kind]; ok {
		return "none"
	}
	return "normal"
}
