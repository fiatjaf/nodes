package helpers

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/awalterschulze/gographviz"
	"log"
	"os/exec"
	"strings"
)

func GenerateDotString(res []struct {
	A    string
	B    string
	REL  string
	URLS []string
}) string {
	// split our slice in a slice with only nodes and a slice with only rels
	var splitAt int
	for index, row := range res {
		if row.B != "" {
			splitAt = index
			break
		}
	}

	graphName := "nodes"
	g := gographviz.NewGraph()
	g.SetName(graphName)
	g.SetDir(true)
	g.AddAttr(graphName, "rankdir", "LR")

	// add nodes
	for _, row := range res[:splitAt] {
		g.AddNode(graphName, fmt.Sprintf("\"%s\"", row.A), map[string]string{
			"label":    nodeTable(row.A, row.URLS),
			"shape":    "box",
			"fontname": "Verdana",
			"fontsize": "9",
		})
	}

	// add edges
	for _, row := range res[splitAt:] {
		g.AddEdge(fmt.Sprintf("\"%s\"", row.A), fmt.Sprintf("\"%s\"", row.B), true, map[string]string{
			"label":    row.REL,
			"dir":      relationshipDir(row.REL),
			"tooltip":  row.REL,
			"fontname": "Verdana",
			"fontsize": "9",
		})
	}

	return g.String()
}

func nodeTable(name string, urls []string) string {
	table := "<<table border=\"0\" cellborder=\"0\" cellpadding=\"4\">"
	table += fmt.Sprintf("<tr><td bgcolor=\"black\" align=\"center\"><font color=\"white\"><b> %s </b></font></td></tr>", name)
	for _, url := range urls {
		table += fmt.Sprintf("<tr><td href=\"%s\" target=\"_blank\" tooltip=\"%s\">%s</td></tr>", url, url, url)
	}
	table += "</table>>"
	return table
}

func relationshipDir(kind string) string {
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

func RenderGraph(dot string) bytes.Buffer {
	var out bytes.Buffer
	cmd := exec.Command("dot", "-Tsvg")
	cmd.Stdin = strings.NewReader(dot)
	cmd.Stdout = &out
	stderr, _ := cmd.StderrPipe()
	go func() {
		in := bufio.NewScanner(stderr)
		for in.Scan() {
			log.Print(in.Text())
		}
	}()

	err := cmd.Run()
	if err != nil {
		log.Print(err.Error())
	}
	return out
}
