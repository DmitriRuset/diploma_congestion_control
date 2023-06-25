package helpers

// Not implemented in Golang
// Max returns the larger of x or y.
func Max(x, y uint32) uint32 {
	if x < y {
		return y
	}
	return x
}

// Min returns the smaller of x or y.
func Min(x, y uint32) uint32 {
	if x > y {
		return y
	}
	return x
}
