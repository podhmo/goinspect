package main

import (
	"os"

	"github.com/podhmo/goinspect/graph"
)

func main() {
	// https://mermaid-js.github.io/mermaid/#/flowchart?id=minimum-length-of-a-link
	g := graph.StringGraph()
	start := g.Madd("Start")
	isIt := g.Madd("is it?")
	isIt.Metadata.Shape = graph.ShapeRhombus
	g.LinkTo(start, isIt)

	ok := g.Madd("OK")
	g.LinkTo(isIt, ok)

	rethink := g.Madd("Rethink")
	g.LinkTo(ok, rethink)
	g.LinkTo(rethink, isIt)

		end := g.Madd("End")
	g.LinkTo(isIt, end)

	graph.RenderMermaid(os.Stdout, g)
}
