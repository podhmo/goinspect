package x

func (w *W) Method(s S) {
	G0()
	defer log()()
	W0{}.M()
	G0()
}

func (w W) String() string { return "W" }

type W struct{}

type W0 struct{}

func (w W0) M() {
	G0()
}
