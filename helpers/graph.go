package helpers

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/awalterschulze/gographviz"
	"html"
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
	// escape name
	name = html.EscapeString(name)

	// turn big things like 'how do we eat bananas' into 'how ... bananas'
	nameshow := name
	if len(name) > 45 {
		nameshow = name[:28] + " ... " + name[len(name)-12:]
	}

	// escape urls
	for i := range urls {
		urls[i] = html.EscapeString(urls[i])
	}

	table := fmt.Sprintf("<<table tooltip=\"%s\" border=\"0\" cellborder=\"0\" cellpadding=\"4\">", name)
	table += fmt.Sprintf("<tr><td bgcolor=\"black\" align=\"center\"><font color=\"white\"><b> %s </b></font></td></tr>", nameshow)
	for _, url := range urls {
		// remove url endings on show
		urlshow := url
		if len(urlshow) > 45 {
			urlshow = urlshow[:42] + "..."
		}
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
