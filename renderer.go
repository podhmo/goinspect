package goinspect

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
	})
	return nil
}