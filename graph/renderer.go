package graph

import (
	"fmt"
	"io"
)

type RenderFunc[K comparable, T any] func(io.Writer, *Graph[K, T]) error

func RenderText[K comparable, T any](w io.Writer, g *Graph[K, T]) error {
	g.WalkPath(func(path []*Node[T]) {
		if len(path) == 1 {
			n := path[0]
			fmt.Fprintf(w, "%v;\n", n.Value)
		} else {
			n, next := path[len(path)-2], path[len(path)-1]
			fmt.Fprintf(w, "%v -> %v;\n", n.Value, next.Value)
		}
	}, nil)
	return nil
}

func RenderMermaid[K comparable, T any](w io.Writer, g *Graph[K, T]) error {
	fmt.Fprintln(w, "```mermaid")
	fmt.Fprintln(w, "flowchart TB")
	g.WalkPath(func(path []*Node[T]) {
		if len(path) == 1 {
			n := path[0]
			switch n.Metadata.Shape {
			case ShapeRhombus:
				fmt.Fprintf(w, "\tG%d{%v};\n", n.ID, n.Value)
			default:
				fmt.Fprintf(w, "\tG%d[%v];\n", n.ID, n.Value)
			}
		} else {
			n, next := path[len(path)-2], path[len(path)-1]
			fmt.Fprintf(w, "\tG%d --> G%d\n", n.ID, next.ID)
		}
	}, nil)
	fmt.Fprintln(w, "```")
	return nil
}

type Shape string

const (
	ShapeText    Shape = ""
	ShapeRhombus Shape = "rhombus"
)
