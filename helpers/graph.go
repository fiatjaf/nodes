package helpers

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/awalterschulze/gographviz"
	"io"
	"log"
	"net/url"
	"os/exec"
	"strconv"
	"strings"
)

func GenerateDotString(res []struct {
	ANAME   string
	BNAME   string
	AID     int
	BID     int
	RELKIND string
	RELID   int
	URLS    []string
}, attrs url.Values) string {
	attrs.Del("url")
	attrs.Del("r")

	// split our slice in a slice with only nodes and a slice with only rels
	var splitAt int
	for index, row := range res {
		if row.BNAME != "" {
			splitAt = index
			break
		}
	}

	graphName := "nodes"
	g := gographviz.NewGraph()
	g.SetName(graphName)
	g.SetDir(true)
	g.AddAttr(graphName, "size", "\"15,1000\"")
	g.AddAttr(graphName, "ratio", "compress")
	for k := range attrs {
		g.AddAttr(graphName, k, attrs.Get(k))
	}

	// add nodes
	for _, row := range res[:splitAt] {
		g.AddNode(graphName, strconv.Itoa(row.AID), map[string]string{
			"id":       strconv.Itoa(row.AID),
			"label":    nodeTable(row.ANAME, row.URLS),
			"shape":    "box",
			"fontname": "Verdana",
			"fontsize": "9",
		})
	}

	// add edges
	for _, row := range res[splitAt:] {
		g.AddEdge(strconv.Itoa(row.AID), strconv.Itoa(row.BID), true, map[string]string{
			"id":       strconv.Itoa(row.RELID),
			"label":    row.RELKIND,
			"dir":      relationshipDir(row.RELKIND),
			"tooltip":  row.RELKIND,
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
		table += fmt.Sprintf("<tr><td href=\"%s\" target=\"_blank\" tooltip=\"%s\">%s</td></tr>", url, url, url[7:])
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
	unflatten := exec.Command("unflatten", "-l 10")
	graphviz := exec.Command("dot", "-Tsvg")

	reader, writer := io.Pipe()

	unflatten.Stdin = strings.NewReader(dot)
	unflatten.Stdout = writer
	graphviz.Stdin = reader
	graphviz.Stdout = &out

	var stderr io.ReadCloser
	stderr, _ = unflatten.StderrPipe()
	go func(stderr io.ReadCloser) {
		in := bufio.NewScanner(stderr)
		for in.Scan() {
			log.Print(in.Text())
		}
	}(stderr)
	stderr, _ = graphviz.StderrPipe()
	go func(stderr io.ReadCloser) {
		in := bufio.NewScanner(stderr)
		for in.Scan() {
			log.Print(in.Text())
		}
	}(stderr)
	unflatten.Start()
	graphviz.Start()
	unflatten.Wait()
	writer.Close()
	graphviz.Wait()
	return out
}
