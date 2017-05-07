package svg

import (
	"fmt"
)

func ExamplePath() {
	segs, _ := ParsePath("M1 2L2 3l0-1z")
	for _, seg := range segs {
		switch s := seg.(type) {
		case Move:
			fmt.Print("M(", s.To.X, s.To.Y, ") ")
		case Line:
			fmt.Print("L(", s.To.X, s.To.Y, ") ")
		case Close:
			fmt.Print("C(", s.To.X, s.To.Y, ") ")
		default:
			panic(s)
		}
	}
	// Output: M(1 2) L(2 3) L(2 2) C(1 2)
}
