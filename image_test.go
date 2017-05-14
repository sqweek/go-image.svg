package svg

import (
	"fmt"
	"image"
	"math"
	"testing"
	"golang.org/x/image/math/fixed"
)

func feq(a, b float64) bool {
	return math.Abs(a - b) < 1e-6
}

func TestFixedPt(t *testing.T) {
	eq := func(s interface{}, a, b fixed.Int26_6) {
		if a != b {
			t.Errorf("%v: %v != %v", s, a, b)
		}
	}
	try := func(f float64, x fixed.Int26_6) {
		eq(f, fix(f), x)
		eq(-f, fix(-f), -x)
	}
	try(42.375, fixed.Int26_6(42<<6 + 1<<4 + 1<<3))
	try(1.5, fixed.Int26_6(1<<6 + 1<<5))
	try(1, fixed.I(1))
	try(0.5, fixed.Int26_6(1<<5))
	try(0.25, fixed.Int26_6(1<<4))
	try(0, fixed.Int26_6(0))
}

func TestTransform(t *testing.T) {
	bounds := Bounds{Min: Point{233, 64}, Max: Point{633, 264}}
	eq := func(p Point, ip image.Point) {
		if !feq(p.X, float64(ip.X)) || !feq(p.Y, float64(ip.Y)) {
			t.Errorf("%v != %v", p, ip)
		}
	}
	try := func(r image.Rectangle) {
		xform := boundsToRect(bounds, r)
		eq(xform.Apply(bounds.Min), r.Min)
		eq(xform.Apply(bounds.Max), r.Max)
	}
	try(image.Rect(0, 0, 400, 200))
	try(image.Rect(0, 0, 200, 100))
	try(image.Rect(-400, -200, 0, 0))
	try(image.Rect(100, 100, 500, 300))
	try(image.Rect(-50, -20, 30, 50))
}

func fmtimg(img image.Image) string {
	s := ""
	x0, y0 := img.Bounds().Min.X, img.Bounds().Min.Y
	for iy := 0; iy < img.Bounds().Dy(); iy++ {
		for ix := 0; ix < img.Bounds().Dx(); ix++ {
			_, _, _, a := img.At(x0 + ix, y0 + iy).RGBA()
			d := (10*a) / 65536
			s += fmt.Sprintf(" %d", d)
		}
		s += "\n"
	}
	return s
}

func TestOffsetImg(t *testing.T) {
	segs, _ := ParsePath("M1 2L2 3l0-1z")
	bounds := PathBounds(segs)
	r1 := image.Rect(-10, -10, 20, 10)
	r2 := image.Rect(0, 0, 30, 20)
	img1 := PathMask(segs, bounds, r1)
	img2 := PathMask(segs, bounds, r2)
	s1 := fmtimg(img1)
	s2 := fmtimg(img2)
	if s1 != s2 {
		t.Errorf("\n%s\n\n%s\n", s1, s2)
	}
}