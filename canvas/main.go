package canvas

import (
	"image"
	"image/color"
	"math"

	"github.com/fogleman/gg"
)

// Canvas represents a drawing surface
type Canvas struct {
	Context *gg.Context
	RGBA    *image.RGBA
}

// NewCanvas creates a new canvas with the specified dimensions
func NewCanvas(width, height int) *Canvas {
	return &Canvas{
		Context: gg.NewContext(width, height),
	}
}

// MoveTo starts a new sub-path at the given (x, y) coordinates
func (c *Canvas) MoveTo(x, y float64) {
	c.Context.MoveTo(x, y)
}

// LineTo adds a straight line to the current path
func (c *Canvas) LineTo(x, y float64) {
	c.Context.LineTo(x, y)
}

// QuadraticBezierTo adds a cubic Bézier curve to the current path
func (c *Canvas) QuadraticBezierTo(cp1x, cp1y, cp2x, cp2y, x, y float64) {
	c.Context.CubicTo(cp1x, cp1y, cp2x, cp2y, x, y)
}

func (c *Canvas) QuadraticBezier(cp1x, cp1y, cp2x, cp2y, x, y float64) {
	// Get the current point
	current, _ := c.Context.GetCurrentPoint()

	// If the current path doesn't exist, create one
	if current.X == 0 && current.Y == 0 {
		c.Context.MoveTo(cp1x, cp1y)
	}

	// Add the quadratic curve
	c.Context.CubicTo(cp1x, cp1y, cp2x, cp2y, x, y)
}

// QuadraticCurveTo adds a quadratic Bézier curve to the current path
func (c *Canvas) QuadraticCurveTo(cpx, cpy, x, y float64) {
	c.Context.QuadraticTo(cpx, cpy, x, y)
}

func (c *Canvas) QuadraticCurve(x1, y1, x, y float64) {
	// Get the current point
	current, _ := c.Context.GetCurrentPoint()

	// If the current path doesn't exist, create one
	if current.X == 0 && current.Y == 0 {
		c.Context.MoveTo(x1, y1)
	}

	// Add the quadratic curve
	c.Context.QuadraticTo(x1, y1, x, y)
}

// ArcTo adds a circular arc to the current path
func (c *Canvas) ArcTo(x1, y1, x2, y2, radius float64) {
	// Get the current point (start of the arc)
	current, _ := c.Context.GetCurrentPoint()

	// Calculate vectors for the current line segment (P0 -> P1) and (P1 -> P2)
	dx1, dy1 := x1-current.X, y1-current.Y
	dx2, dy2 := x2-x1, y2-y1

	// Normalize the vectors
	len1 := math.Hypot(dx1, dy1)
	len2 := math.Hypot(dx2, dy2)

	// Handle degenerate cases (collinear points or radius too small)
	if len1 == 0 || len2 == 0 || radius == 0 {
		// If degenerate, just draw a line to the first point
		c.Context.LineTo(x1, y1)
		return
	}

	// Unit vectors
	ux1, uy1 := dx1/len1, dy1/len1
	ux2, uy2 := dx2/len2, dy2/len2

	// Angle between the two vectors
	cosAngle := ux1*ux2 + uy1*uy2
	angle := math.Acos(cosAngle)

	// Compute the distance from P1 to the arc's center
	tanHalfAngle := math.Tan(angle / 2)
	distance := radius / tanHalfAngle

	// Calculate the arc's center
	centerX := x1 - ux1*distance
	centerY := y1 - uy1*distance

	// Calculate start and end angles for the arc
	startAngle := math.Atan2(current.Y-centerY, current.X-centerX)
	endAngle := math.Atan2(y1-centerY, x1-centerX)

	// Determine the direction of the arc (clockwise or anticlockwise)
	// anticlockwise := (ux1*uy2 - uy1*ux2)

	// Add the arc to the path
	c.Arc(centerX, centerY, radius, startAngle, endAngle)
}

// Arc draws an arc on the current path
func (c *Canvas) Arc(x, y, radius, startAngle, endAngle float64) {
	c.Context.DrawArc(x, y, radius, startAngle, endAngle)
}

// Rect creates a rectangular path
func (c *Canvas) Rect(x, y, width, height float64) {
	c.Context.DrawRectangle(x, y, width, height)
}

// Fill fills the current path with the current fill style
func (c *Canvas) Fill() {
	c.Context.Fill()
}

// Stroke strokes the current path with the current stroke style
func (c *Canvas) Stroke() {
	c.Context.Stroke()
}

// BeginPath starts a new path
func (c *Canvas) Clip() {
	c.Context.Clip()
}

// BeginPath starts a new path
func (c *Canvas) BeginPath() {
	c.Context.NewSubPath()
}

// ClosePath closes the current path
func (c *Canvas) ClosePath() {
	c.Context.ClosePath()
}

// SetFillStyle sets the fill color
func (c *Canvas) SetFillStyle(r, g, b, a uint8) {
	c.Context.SetRGBA(float64(r)/255, float64(g)/255, float64(b)/255, float64(a)/255)
}

// SetStrokeStyle sets the stroke color
func (c *Canvas) SetStrokeStyle(r, g, b, a uint8) {
	c.Context.SetStrokeStyle(gg.NewSolidPattern(color.RGBA{R: r, G: g, B: b, A: a}))
}

func (c *Canvas) SetLineDash(dashes ...float64) {
	c.Context.SetDash(dashes...)
}

// ClearRect clears the specified rectangle area
func (c *Canvas) ClearRect(x, y, width, height float64) {
	c.Context.SetRGBA(1, 1, 1, 1) // Assuming white background
	c.Context.DrawRectangle(x, y, width, height)
	c.Context.Fill()
}

// FillRect fills a rectangle with the current fill style
func (c *Canvas) FillRect(x, y, width, height float64) {
	c.Rect(x, y, width, height)
	c.Fill()
}

// StrokeRect strokes a rectangle with the current stroke style
func (c *Canvas) StrokeRect(x, y, width, height float64) {
	c.Rect(x, y, width, height)
	c.Stroke()
}

// Save saves the current drawing state
func (c *Canvas) Save() {
	c.Context.Push()
}

// Restore restores the last saved drawing state
func (c *Canvas) Reset() {
	c.Context.Pop()
	c.Context.ResetClip()
}

// Translate moves the canvas origin to (x, y)
func (c *Canvas) Translate(x, y float64) {
	c.Context.Translate(x, y)
}

// Scale scales the canvas
func (c *Canvas) Scale(sx, sy float64) {
	c.Context.Scale(sx, sy)
}

// Rotate rotates the canvas
func (c *Canvas) Rotate(angle float64) {
	c.Context.Rotate(angle)
}

func (c *Canvas) Ellipse(x, y, radiusX, radiusY, rotation, startAngle, endAngle float64, anticlockwise bool) {
	c.Save()
	c.Translate(x, y)
	c.Rotate(rotation)
	c.Scale(radiusX, radiusY)
	c.Arc(0, 0, 1, startAngle, endAngle)
	c.Reset()
}
func (c *Canvas) MeasureText(s string) (w, h float64) {
	return c.Context.MeasureString(s)
}
func (c *Canvas) FillText(text string, x, y float64) {
	c.Context.DrawStringAnchored(text, x, y, 0, 0)
	c.Context.Fill()
}
func (c *Canvas) StrokeText(text string, x, y float64) {
	c.Context.DrawStringAnchored(text, x, y, 0, 0)
	c.Context.Stroke()
}
func (c *Canvas) SetFont(fontPath string, fontSize float64) error {
	// !ISSUE: Breaks the fs rulesz
	font, err := gg.LoadFontFace(fontPath, fontSize)
	if err != nil {
		return err
	}
	c.Context.SetFontFace(font)
	return nil
}
func (c *Canvas) SetLineWidth(width float64) {
	c.Context.SetLineWidth(width)
}
func (c *Canvas) SetLineCap(cap string) {
	switch cap {
	case "butt":
		c.Context.SetLineCap(gg.LineCapButt)
	case "round":
		c.Context.SetLineCap(gg.LineCapRound)
	case "square":
		c.Context.SetLineCap(gg.LineCapSquare)
	}
}

//	func (c *Canvas) SetLineJoin(join string) {
//		switch join {
//		case "miter":
//			c.Context.SetLineJoin(gg.LineJoinMiter)
//		case "round":
//			c.Context.SetLineJoin(gg.LineJoinRound)
//		case "bevel":
//			c.Context.SetLineJoin(gg.LineJoinBevel)
//		}
//	}
func (c *Canvas) SetGlobalAlpha(alpha float64) {
	c.Context.SetRGBA(1, 1, 1, alpha)
}

func (c *Canvas) LoadImage(path string) image.Image {
	i, _ := gg.LoadImage(path)
	return i
}

func (c *Canvas) DrawImage(img image.Image, x, y float64) {
	c.Context.DrawImage(img, int(x), int(y))
}

//	func (c *Canvas) SetTransform(a, b, c1, d, e, f float64) {
//		c.Context.Identity()
//		c.Context.Transform(a, b, c1, d, e, f)
//	}
//
//	func (c *Canvas) GetImageData(x, y, width, height int) *image.Image {
//		img := c.Context.Image()
//		subImage := img.SubImage(image.Rect(x, y, x+width, y+height))
//		return subImage.(*image.Image)
//	}
func (c *Canvas) PutImageData(imgData image.Image, x, y int) {
	bounds := imgData.Bounds()
	for i := bounds.Min.X; i < bounds.Max.X; i++ {
		for j := bounds.Min.Y; j < bounds.Max.Y; j++ {
			r, g, b, a := imgData.At(i, j).RGBA()
			c.SetFillStyle(uint8(r), uint8(g), uint8(b), uint8(a))
			c.Context.SetPixel(x+i, y+j)
		}
	}
}

// RoundedRect draws a rectangle with rounded corners.
// The `radii` parameter is a slice of four float64 values representing the radius for each corner:
// [top-left, top-right, bottom-right, bottom-left]. If a radius is 0, the corner will be square.
func (c *Canvas) RoundedRect(x, y, width, height float64, radii []float64) {
	if len(radii) != 4 {
		panic("radii must contain exactly 4 values")
	}

	// Clamp each radius to avoid overlap
	for i := range radii {
		radii[i] = math.Min(radii[i], math.Min(width/2, height/2))
	}

	// Extract radii for each corner
	topLeft := radii[0]
	topRight := radii[1]
	bottomRight := radii[2]
	bottomLeft := radii[3]

	// Start at the top-left corner
	c.Context.MoveTo(x+topLeft, y)

	// Top edge
	c.Context.LineTo(x+width-topRight, y)
	// Top-right corner
	if topRight > 0 {
		c.Arc(x+width-topRight, y+topRight, topRight, -math.Pi/2, 0)
	}

	// Right edge
	c.Context.LineTo(x+width, y+height-bottomRight)
	// Bottom-right corner
	if bottomRight > 0 {
		c.Arc(x+width-bottomRight, y+height-bottomRight, bottomRight, 0, math.Pi/2)
	}

	// Bottom edge
	c.Context.LineTo(x+bottomLeft, y+height)
	// Bottom-left corner
	if bottomLeft > 0 {
		c.Arc(x+bottomLeft, y+height-bottomLeft, bottomLeft, math.Pi/2, math.Pi)
	}

	// Left edge
	c.Context.LineTo(x, y+topLeft)
	// Top-left corner
	if topLeft > 0 {
		c.Arc(x+topLeft, y+topLeft, topLeft, math.Pi, 3*math.Pi/2)
	}

	// Close the path
	c.Context.ClosePath()
}

// !TODO: Make own gradient type that applies to gg to make portable
func (c *Canvas) CreateLinearGradient(x0, y0, x1, y1 float64) gg.Gradient {
	return gg.NewLinearGradient(x0, y0, x1, y1)
}
