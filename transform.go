package svg

import (
	"image"
	"math"
	"golang.org/x/image/math/fixed"
)

type Bounds struct {
	Min, Max Point
	hasPoints bool
}

func PathBounds(path []Segment) (b Bounds) {
	for _, seg := range path {
		b.segment(seg)
	}
	return b
}

func (b *Bounds) segment(segi Segment) {
	switch seg := segi.(type) {
	case Move:
		b.GrowPt(seg.To)
	case Line:
		b.GrowPt(seg.To)
	case Quadratic:
		b.GrowPt(seg.Ctl, seg.To)
	case Cubic:
		b.GrowPt(seg.Ctl1, seg.Ctl2, seg.To)
	case Arc:
		b.GrowPt(seg.To) // TODO account for arc
	case Close: // do nothing
	default:
		panic(seg)
	}
}

func update(min, max *float64, val float64) {
	if val < *min {
		*min = val
	} else if val > *max {
		*max = val
	}
}

func (b *Bounds) GrowPt(pts ...Point) {
	for _, pt := range pts {
		if !b.hasPoints {
			b.Min = pt
			b.Max = pt
			b.hasPoints = true
			continue
		}
		update(&b.Min.X, &b.Max.X, pt.X)
		update(&b.Min.Y, &b.Max.Y, pt.Y)
	}
}

func (b Bounds) Dx() float64 {
	return b.Max.X - b.Min.X
}

func (b Bounds) Dy() float64 {
	return b.Max.Y - b.Min.Y
}

func (b Bounds) Inset(n float64) Bounds {
	return b.Border(-n, -n, -n, -n)
}

func (b Bounds) Border(left, top, right, bottom float64) Bounds {
	out := b
	out.Min.X -= left
	out.Min.Y -= top
	out.Max.X += right
	out.Max.Y += bottom
	return out
}

func (b Bounds) WidthForHeight(desiredHeight float64) float64 {
	ratio := desiredHeight / b.Dy()
	return b.Dx() * ratio
}

func (b Bounds) HeightForWidth(desiredWidth float64) float64 {
	ratio := desiredWidth / b.Dx()
	return b.Dy() * ratio
}

type Transform struct {
	Trans, Scale Point
}

func boundsToRect(b Bounds, r image.Rectangle) (t Transform) {
	t.Scale.X = float64(r.Dx()) / b.Dx()
	t.Scale.Y = float64(r.Dy()) / b.Dy()
	t.Trans.X = float64(r.Min.X) - t.Scale.X*b.Min.X
	t.Trans.Y = float64(r.Min.Y) - t.Scale.Y*b.Min.Y
	return
}

func (t Transform) Apply(p Point) Point {
	return Point{
		X: p.X * t.Scale.X + t.Trans.X,
		Y: p.Y * t.Scale.Y + t.Trans.Y,
	}
}

func (t Transform) fix(p Point) fixed.Point26_6 {
	p = t.Apply(p)
	return fixed.Point26_6{X: fix(p.X), Y: fix(p.Y)}
}

func fix(num float64) fixed.Int26_6 {
	if num < 0 {
		return -fixpos(-num)
	}
	return fixpos(num)
}

func fixpos(num float64) fixed.Int26_6 {
	i, f := math.Modf(num)
	n := int(f * 64 + 0.5)

	return fixed.Int26_6((n & 0x3f) | ((int(i) & 0x03ffffff) << 6))
}
