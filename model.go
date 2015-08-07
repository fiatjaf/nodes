package main

type Neo4jNode interface {
	Exists() bool
}

type URL struct {
	Url   string
	Title string
}

func (o URL) Exists() bool {
	return o.Url != ""
}

type Node struct {
	Name string
}

func (o Node) Exists() bool {
	return o.Name != ""
}
