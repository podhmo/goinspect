package x

func R(n int) int {
	if n <= 0 {
		H()
		return 1
	} else {
		return R(n-1) * n
	}
}

func Odd(n int) bool {
	if n == 1 {
		H()
		return false
	} else {
		return Even(n - 1)
	}
}
func Even(n int) bool {
	if n == 0 {
		H()
		return false
	} else {
		return Odd(n - 1)
	}
}

func RecRoot(n int) {
	R(n)
	Odd(n)
}
