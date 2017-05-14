package svg

import (
	"image"

	"github.com/golang/freetype/raster"
	"golang.org/x/image/math/fixed"
)

/* PathMask paints a series of Segments as an alpha-mask image */
func PathMask(path []Segment, b Bounds, r image.Rectangle) *image.Alpha {
	// r is aligned to (0, 0) for the sake of the raster package, which implicitly works
	// in a (0, 0)-(w, h) coordinate space. The Dx/Dy fields are set before rasterizing
	// to translate into the caller's desired coordinate space.
	xform := boundsToRect(b, r.Sub(r.Min))
	pt := func(p Point)fixed.Point26_6 { return xform.fix(p) }
	var rpath raster.Path
	for _, segi := range path {
		switch seg := segi.(type) {
		case Move:
			rpath.Start(pt(seg.To))
		case Close:
			rpath.Add1(pt(seg.To))
		case Line:
			rpath.Add1(pt(seg.To))
		case Quadratic:
			rpath.Add2(pt(seg.Ctl), pt(seg.To))
		case Cubic:
			rpath.Add3(pt(seg.Ctl1), pt(seg.Ctl2), pt(seg.To))
		// case Arc: // TODO add two quadratics?
		default:
			panic("unimplemented")
		}
	}
	mask := image.NewAlpha(r)
	ras := raster.NewRasterizer(r.Dx(), r.Dy())
	ras.Dx, ras.Dy = r.Min.X, r.Min.Y
	ras.AddPath(rpath)
	ras.Rasterize(raster.NewAlphaSrcPainter(mask))
	return mask
}
