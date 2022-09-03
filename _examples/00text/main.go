package main

import (
	"os"

	"github.com/podhmo/goinspect"
)

func main() {
	// https://mermaid-js.github.io/mermaid/#/flowchart?id=minimum-length-of-a-link
	g := goinspect.StringGraph()
	start, _ := g.Add("Start")
	isIt, _ := g.LinkTo(start, "Is it?")
	isIt.Metadata.Shape = goinspect.ShapeRhombus
	ok, _ := g.LinkTo(isIt, "OK")
	rethink, _ := g.LinkTo(ok, "Rethink")
	g.LinkTo(rethink, "Is it?")

	g.LinkTo(isIt, "End")

	goinspect.RenderMermaid(os.Stdout, g)
}
