package main

import (
	"fmt"
	"math"
	"image"
	"image/color"
	"image/draw"
	"time"

	"github.com/skelterjohn/go.wde"
	_ "github.com/skelterjohn/go.wde/init"
	"github.com/sqweek/go-image.svg"
)

var bassstr = "m190.85 451.25c11.661 14.719 32.323 24.491 55.844 24.491 36.401 0 65.889-23.372 65.889-52.214s-29.488-52.214-65.889-52.214c-20.314 4.1522-28.593 9.0007-33.143-2.9091 17.976-54.327 46.918-66.709 96.546-66.709 65.914 0 96.969 59.897 96.969 142.97-18.225 190.63-205.95 286.75-246.57 316.19 5.6938 13.103 5.3954 12.631 5.3954 12.009 189.78-86.203 330.69-204.43 330.69-320.74 0-92.419-58.579-175.59-187.72-172.8-77.575 0-170.32 86.203-118 171.93zm328.1-89.88c0 17.852 14.471 32.323 32.323 32.323s32.323-14.471 32.323-32.323-14.471-32.323-32.323-32.323-32.323 14.471-32.323 32.323zm0 136.75c0 17.852 14.471 32.323 32.323 32.323s32.323-14.471 32.323-32.323-14.471-32.323-32.323-32.323-32.323 14.471-32.323 32.323z"

var treblestr = "m278.89 840.5c0 23.811 27.144 51.133 48.193 55.108 0 0 22.508 5.6134 41.03 1.1224 3.3461 0.75146 59.498-9.5416 60.62-84.755l-11.843-91.536c7.8919-1.3481 33.643-13.763 41.592-18.478 21.812-12.943 31.355-31.636 39.852-45.465 9.5416-20.768 15.155-34.946 15.155-57.556 0-57.185-46.341-103.54-103.5-103.54-8.2386 0-16.243 0.97722-23.934 2.7962l-10.216-73.464c38.684-43.006 77.897-95.869 91.065-144.31 12.91-63.989 6.7357-107.21-27.504-125.73-31.432-7.2969-69.589 35.643-92.052 97.666-14.594 33.116 3.9976 96.385 12.349 122.36-61.528 58.983-122.74 101.28-129.1 203.51 0 80.053 64.874 144.94 144.89 144.94 7.6452 0 20.275 0.179 34.643-1.6835l6.7583 91.412s2.593 66.862-48.293 75.922c-20.196 3.5928-39.674 7.7565-34.07-5.6118 15.739-8.1967 34.07-21.263 34.07-42.704 0-26.683-20.085-48.316-44.849-48.316-24.774 0-44.859 21.633-44.859 48.316zm90.975-409.14 9.5303 68.634c-42.22 13.057-72.901 52.426-72.901 98.946 0 52.649 62.945 77.202 61.239 64.706-24.54-10.587-33.633-22.486-33.633-50.911 0-32.129 21.958-59.115 51.673-66.806l22.665 163.41c-5.0974 1.3804-10.598 2.5479-16.558 3.4799-86.249-1.7287-126.79-47.632-126.79-127.68 0.002-44.846 52.809-96.576 104.77-153.77zm45.197 275.97-22.599-162.78c3.4574-0.5386 7.0163-0.83047 10.632-0.83047 38.1 0 68.994 30.905 68.994 69.017-0.76275 30.67-2.2237 76.336-57.027 94.59zm-48.06-342.71c-15.548-52.459-1.4932-78.603 7.6323-98.699 21.565-47.452 40.537-64.706 59.857-80.108 25.124 9.8561 13.784 80.108 10.012 80.108-11.831 41.446-54.276 74.395-77.502 98.699z"

var tristr = "m33 33h33l-16.7 33z"

func main() {
	go func() {
		wdemain()
		wde.Stop()
	}()
	wde.Run()
}

func wdemain() {
	chk := func(err error) { if err != nil {panic(err)} }
	path, err := svg.ParsePath(treblestr)
	chk(err)
	bounds := svg.PathBounds(path)
	fmt.Println("Bounds: ", bounds.Dx(), bounds.Dy(), bounds.Dx() / bounds.Dy())
	w, err := wde.NewWindow(400, 200)
	chk(err)

	w.SetTitle("SVG test")
	w.Show()

	var refresh *time.Timer
	events: for ei := range w.EventChan() {
		switch e := ei.(type) {
		case wde.MouseUpEvent:
			fmt.Println(e.Where)
		case wde.ResizeEvent:
			if refresh != nil {
				refresh.Stop()
			}
			refresh = time.AfterFunc(50*time.Millisecond, func() {render(path, bounds, w)})
		case wde.CloseEvent:
			break events
		}
	}
}

func render(path []svg.Segment, pBounds svg.Bounds, win wde.Window) {
	screen := win.Screen()
	bounds := screen.Bounds()
	w := bounds.Dx() - 20
	h := int(math.Ceil(pBounds.HeightForWidth(float64(w))))
	if h > bounds.Dy() {
		h = bounds.Dy() -20
		w = int(math.Ceil(pBounds.WidthForHeight(float64(h))))
	}
	fmt.Println("resized to ", w, h)
	img := image.NewRGBA(bounds)
	draw.Draw(img, img.Bounds(), &image.Uniform{color.RGBA{0xff, 0, 0, 0xff}}, image.ZP, draw.Src)
	pts := []image.Point{image.Pt(-15, -20), image.Pt(0, 0), image.Pt(15, 20)}
	var masks []*image.Alpha
	for i, delta := range pts {
		mr := image.Rect(0, 0, w, h).Add(delta)
		mask := svg.PathMask(path, pBounds, mr)
		masks = append(masks, mask)
		rMin := image.Pt(10 + i * (w + 10), 10)
		r := image.Rectangle{rMin, rMin.Add(image.Pt(w, h))}
		draw.Draw(img, r, &image.Uniform{color.White}, image.ZP, draw.Src)
		draw.DrawMask(img, r, &image.Uniform{color.Black}, image.ZP, mask, mr.Min, draw.Over)
	}
	screen.CopyRGBA(img, img.Bounds())
	win.FlushImage()
}
