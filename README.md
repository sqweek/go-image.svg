# go-image.svg
Currently allows an SVG path to be rendered to a standard image.Alpha.
The package is named so that it is referred to as `svg` following an `import "github.com/sqweek/go-image.svg"`,
much as `import "image/png"` gives you a package referred to as `png`.

# example (render an SVG path to PNG)

    package main
    
    import (
        "image"
        "image/color"
        "image/draw"
        "image/png"
        "os"
        "github.com/sqweek/go-image.svg"
    )
    
    func main() {
        // Parse SVG path data into abstract representation
        path, err := svg.ParsePath("m33 33h33l-16.7 33z")
        if err != nil {
            panic(err)
        }
        // Calculate approx. bounds of the path
        bounds := svg.PathBounds(path)
        // A border can be added with bounds.Inset(-n) or bounds.Border(left, top, right, bot)
        // Note that n/left/top/right/bot are in the same coordinates as the SVG path

        // Render the path to an image.Alpha
        r := image.Rect(0, 0, 200, 200)
        mask := svg.PathMask(path, bounds, r)

        // Prepare to export a PNG
        img := image.NewRGBA(r)
        bg, fg := color.White, color.Black // background/foreground colours
        draw.Draw(img, r, &image.Uniform{bg}, image.ZP, draw.Src)
        draw.DrawMask(img, r, &image.Uniform{fg}, image.ZP, mask, image.ZP, draw.Over)

        f, err := os.Create("tmp.png")
        if err != nil {
            panic(err)
        }
        defer f.Close()
        err = png.Encode(f, img)
        if err != nil {
            panic(err)
        }
    }

# bugs

1. SVG Elliptical Arc segments cannot be rasterised (yet)
2. XML not supported at all (yet)
3. `svg.ParsePath` should probably accept a `Point` argument for when the first command in the path is relative (ie. not absolute)
4. There should be a way to render a path to an existing `Image` rather than always allocating a new one
5. It should be possible to programatically build a path without parsing a string


