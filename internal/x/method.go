package x

func (w *W) Method(s S) {
	G0()
	defer log()()
	G0()
}

func (w W) String() string { return "W" }

type W struct{}
