package x

import "github.com/podhmo/goinspect/internal/x/sub"

type S struct{}

func F(s S) {
	defer log()()
	F0()
	H()
}
func F0() {
	defer log()()
	F1()
}

func F1() {
	defer log()()
	println("F")
	H()
}

func G() {
	defer log()()
	G0()
	sub.X()
}

func G0() {
	defer log()()
	println("G")
	H()
}

func H() {
	println("H")
}

func log() func() {
	println("start")
	return func() {
		println("end")
	}
}
