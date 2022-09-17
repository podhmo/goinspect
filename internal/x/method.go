package x

type state struct{}

func (s *state) mark() {}
func (s *state) eval(v interface{}) {
	s.mark()
}

func (w *W) Method(s S) {
	G0()
	defer log()()
	G0()
}
func (w *W) MethodWithCompoliteLiteral(s S) {
	defer log()()
	W0{}.M0()
	(&W0{}).M1()
}
func (w *W) MethodWithMethodInvoke(s S) {
	defer log()()
	w0 := &W0{}
	w0.M1()
}
func (w *W) MethodWithFactoryFunction(s S) {
	defer log()()
	NewW0().M1()
}

func (w W) String() string { return "W" }

type W struct{}

type W0 struct{}

func (w W0) M0() {
	G0()
	w.Inner()
}
func (w *W0) M1() {
	F0()
	w.Inner()
}
func (w *W0) Inner() {
}
func (w *W0) M2(v interface{}) {
	(&state{}).eval(v)
}
func NewW0() *W0 {
	return &W0{}
}
