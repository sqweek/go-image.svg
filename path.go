package svg

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Segment interface {
	segment_tag()  // no-op; just an interface tag
}

type Point struct {
	X, Y float64
}

// Move represents a repositioning of the pen
type Move struct { To Point }
func (m Move) segment_tag() {}

// Line is a linear segment
type Line struct { To Point }
func (l Line) segment_tag() {}

// Close is a segment closing the current path
type Close struct { To Point }
func (c Close) segment_tag() {}

// A Quadratic bézier curve segment
type Quadratic struct { Ctl, To Point }
func (q Quadratic) segment_tag() {}

// A Cubic bézier curve segment
type Cubic struct { Ctl1, Ctl2, To Point }
func (c Cubic) segment_tag() {}

// An elliptical Arc curve segment (is this needed? could be two Quadratics?)
type Arc struct {
	rx, ry, rot float64
	large_arc, sweep bool
	To Point
}
func (a Arc) segment_tag() {}

func ParsePath(s string) (segs []Segment, err error) {
	p := &PathParser{s: s, orig: s}
	for {
		seg, err := p.Next()
		if err != nil {
			return nil, err
		}
		if seg == nil {
			break
		}
		segs = append(segs, seg)
	}
	return
}

type PathParser struct {
	s string    // the remaining data to be parsed
	cmd rune    // the current curve command character
	n int       // the number of float arguments cmd requires
	pt Point    // current position (used for relative coordinates)
	start Point // start of current subpath
	orig string // original string (for errors)
	prev Segment
}

type ParseError struct {
	Full string // the [entire] string which failed parsing
	Index int   // the index at which parsing failed
	Msg string  // error details
}

func (e ParseError) Context(chars int) string {
	start, end := e.Index - chars, e.Index + chars
	m := ""
	if start < 0 {
		m += e.Full[:e.Index]
	} else {
		m += "…" + e.Full[start:e.Index]
	}
	m += "¡" + e.Full[e.Index:e.Index+1] + "!"
	if e.Index < len(e.Full) - 1 {
		if end > len(e.Full) - 1 {
			m += e.Full[e.Index+1:]
		} else {
			m += e.Full[e.Index+1:end] + "…"
		}
	}
	return m
}

func (e ParseError) Error() string {
	m := "parse error at index " + strconv.Itoa(e.Index) + ": " + e.Context(20)
	if e.Msg != "" {
		m += ": " + e.Msg
	}
	return m
}

func (p PathParser) error(msg string) ParseError {
	return ParseError{p.orig, strings.LastIndex(p.orig, p.s), msg}
}

func (p *PathParser) Next() (seg Segment, err error) {
	p.s = strings.TrimLeft(p.s, " ,")
	if p.s == "" {
		return nil, nil
	}
	first, w1 := utf8.DecodeRuneInString(p.s)
	if unicode.IsLetter(first) {
		p.s = p.s[w1:]
		p.cmd = first
		switch unicode.ToLower(first) {
		case 'z':
			p.n = 0
			return p.makeSeg([]float64{}), nil
		case 'h', 'v':
			p.n = 1
		case 'm', 'l', 't':
			p.n = 2
		case 's', 'q':
			p.n = 4
		case 'c':
			p.n = 6
		case 'a':
			p.n = 7
		default:
			return nil, fmt.Errorf("unrecognised curve command '%c'", first)
		}
	} else if p.cmd == '\000' {
		return nil, fmt.Errorf("path string doesn't begin with a command")
	}

	args := make([]float64, p.n)
	for i := 0; i < p.n; i++ {
		var num string
		offset := 0
		p.s = strings.TrimLeft(p.s, " ,")  // skip whitespace/commas
		if p.s[0] == byte('-') {
			offset = 1
		}
		endpos := offset + strings.IndexFunc(p.s[offset:], func(r rune)bool { return !strings.ContainsRune("0123456789.", r) })
		if endpos < offset {
			endpos = len(p.s)
		}
		num, p.s = p.s[:endpos], p.s[endpos:]
		f, err := strconv.ParseFloat(num, 64)
		if err != nil {
			return nil, p.error(err.Error())
		}
		args[i] = f
	}

	seg = p.makeSeg(args)
	p.prev = seg
	return
}

func (p *PathParser) makeSeg(a []float64) (seg Segment) {
	var dest Point
	switch unicode.ToLower(p.cmd) {
	case 'z':
		return Close{To: p.start}
	case 'h':
		dest = Point{p.resolve(a[0], 0).X, p.pt.Y}
		seg = Line{To: dest}
	case 'v':
		dest = Point{p.pt.X, p.resolve(0, a[0]).Y}
		seg = Line{To: dest}
	case 'l':
		dest = p.resolve(a[0], a[1])
		seg = Line{To: dest}
	case 'm':
		dest = p.resolve(a[0], a[1])
		p.start = dest
		seg = Move{To: dest}
	case 'q':
		ctl := p.resolve(a[0], a[1])
		dest = p.resolve(a[2], a[3])
		seg = Quadratic{Ctl: ctl, To: dest}
	case 't':
		dest = p.resolve(a[0], a[1])
		ctl := p.pt
		if prevq, ok := p.prev.(Quadratic); ok {
			ctl = reflect(prevq.Ctl, p.pt)
		}
		seg = Quadratic{Ctl: ctl, To: dest}
	case 'c':
		ctl1 := p.resolve(a[0], a[1])
		ctl2 := p.resolve(a[2], a[3])
		dest = p.resolve(a[4], a[5])
		seg = Cubic{ctl1, ctl2, dest}
	case 's':
		ctl2 := p.resolve(a[0], a[1])
		dest = p.resolve(a[2], a[3])
		ctl1 := p.pt
		if prevc, ok := p.prev.(Cubic); ok {
			ctl1 = reflect(prevc.Ctl2, p.pt)
		}
		seg = Cubic{ctl1, ctl2, dest}
	case 'a':
		dest = p.resolve(a[5], a[6])
		seg = Arc{rx: a[0], ry: a[1], rot: a[2], large_arc: a[3] != 0, sweep: a[4] != 0, To: dest}
	default:
		// Parser.Next should have returned an error instead
		panic(fmt.Sprintf("unhandled case '%c'", p.cmd))
	}
	p.pt = dest
	return seg
}

// Returns a point, resolving relative coordinates if necessary
func (p *PathParser) resolve(x, y float64) Point {
	if unicode.IsUpper(p.cmd) {
		return Point{x, y} // absolute coordinates
	}
	return Point{p.pt.X + x, p.pt.Y + y}
}

func reflect(p, origin Point) Point {
	dx, dy := p.X - origin.X, p.Y - origin.Y
	return Point{origin.X - dx, origin.Y - dy}
}
