package font

import (
	"grim/element"
	"image"
	"image/color"
	"image/draw"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// LoadSystemFont loads a font from the system fonts directory or loads a specific font by name
func GetFontPath(fontName string, bold int, italic bool) string {
	if len(fontName) == 0 {
		fontName = "serif"
	}

	fonts := strings.Split(fontName, ",")
	for _, font := range fonts {
		font = strings.TrimSpace(font)
		fontPath := tryLoadSystemFontWithBestWeight(font, bold, italic)

		if fontPath != "" {
			return fontPath
		}

		// Check special font families only if it's the first font in the list
		switch font {
		case "sans-serif":
			fontPath = tryLoadSystemFontWithBestWeight("Arial", bold, italic)
		case "monospace":
			fontPath = tryLoadSystemFontWithBestWeight("Andale Mono", bold, italic)
		case "serif":
			fontPath = tryLoadSystemFontWithBestWeight("Georgia", bold, italic)
		}

		if fontPath != "" {
			return fontPath
		}

	}

	// Default to serif if none of the specified fonts are found
	return tryLoadSystemFontWithBestWeight("Georgia", bold, italic)
}

func tryLoadSystemFontWithBestWeight(fontName string, desiredWeight int, italic bool) string {
	font := fontName
	if italic {
		font += " Italic"
	}
	slash := "/"

	if runtime.GOOS == "windows" {
		slash = "\\"
	}

	availableWeights := getAvailableWeights(fontName, italic)
	closestWeight := findClosestWeight(availableWeights, desiredWeight)

	// Get the closest weight name and attempt to load the font
	weightName := getWeightName(closestWeight)
	fontWithWeight := fontName + " " + weightName
	fontWithWeight = strings.TrimSpace(fontWithWeight)
	if italic {
		fontWithWeight += " Italic"
	}

	for _, v := range allFonts {
		if strings.Contains(strings.ToLower(v), strings.ToLower(slash+fontWithWeight)) {
			return v
		}
	}

	return ""
}

func getAvailableWeights(fontName string, italic bool) []int {
	weightMappings := []int{100, 200, 300, 400, 500, 600, 700, 800, 900}
	availableWeights := []int{}

	for _, weight := range weightMappings {
		weightName := getWeightName(weight)
		fontWithWeight := strings.TrimSpace(fontName + " " + weightName)
		if italic {
			fontWithWeight += " Italic"
		}
		if tryLoadFontExists(fontWithWeight) {
			availableWeights = append(availableWeights, weight)
		}
	}

	return availableWeights
}

func tryLoadFontExists(fontWithWeight string) bool {
	slash := "/"
	if runtime.GOOS == "windows" {
		slash = "\\"
	}

	for _, v := range allFonts {
		if strings.Contains(strings.ToLower(v), strings.ToLower(slash+fontWithWeight)) {
			return true
		}
	}

	return false
}

func findClosestWeight(availableWeights []int, desiredWeight int) int {
	if len(availableWeights) == 0 {
		return 400 // Default to "Regular" if no weights are available
	}

	closest := availableWeights[0]
	for _, weight := range availableWeights {
		if abs(weight-desiredWeight) < abs(closest-desiredWeight) {
			closest = weight
		}
	}
	return closest
}

func getWeightName(weight int) string {
	switch weight {
	case 100:
		return "Thin"
	case 200:
		return "ExtraLight"
	case 300:
		return "Light"
	case 400:
		return ""
	case 500:
		return "Medium"
	case 600:
		return "SemiBold"
	case 700:
		return "Bold"
	case 800:
		return "ExtraBold"
	case 900:
		return "Black"
	default:
		return ""
	}
}

var allFonts = getSystemFonts()

func sortByLength(strings []string) {
	sort.Slice(strings, func(i, j int) bool {
		return len(strings[i]) < len(strings[j])
	})
}

func GetFontSize(css map[string]string) float32 {
	fL := len(css["font-size"])

	var fs float32 = 16

	if fL > 0 {
		if css["font-size"][fL-2:] == "px" {
			fs64, _ := strconv.ParseFloat(css["font-size"][0:fL-2], 32)
			fs = float32(fs64)
		}
		if css["font-size"][fL-2:] == "em" {
			fs64, _ := strconv.ParseFloat(css["font-size"][0:fL-2], 32)
			fs = float32(fs64)
		}
	}

	return fs
}

func LoadFont(fontName string, fontSize int, bold int, italic bool) (font.Face, error) {
	// Use a TrueType font file for the specified font name
	fontFile := GetFontPath(fontName, bold, italic)

	// Read the font file
	fontData, err := os.ReadFile(fontFile)
	if err != nil {
		return nil, err
	}

	// Parse the TrueType font data
	fnt, err := truetype.Parse(fontData)
	if err != nil {
		return nil, err
	}

	options := truetype.Options{
		Size:    float64(fontSize),
		DPI:     72,
		Hinting: font.HintingNone,
	}

	// Create a new font face with the specified size
	return truetype.NewFace(fnt, &options), nil
}

func MeasureText(t *element.Text, text string) int {
	var width fixed.Int26_6

	for _, runeValue := range text {
		if runeValue == ' ' {
			// Handle spaces separately, add word spacing
			width += fixed.I(t.WordSpacing)
		} else {
			fnt := *t.Font
			adv, ok := fnt.GlyphAdvance(runeValue)
			if !ok {
				continue
			}

			// Update the total width with the glyph advance and bounds
			width += adv + fixed.I(t.LetterSpacing)
		}
	}

	return width.Round()
}

func MeasureSpace(t *element.Text) int {
	fnt := *t.Font
	adv, _ := fnt.GlyphAdvance(' ')
	return adv.Round()
}

func getSystemFonts() []string {
	var fontPaths []string

	switch runtime.GOOS {
	case "windows":
		fontPaths = append(fontPaths, getWindowsFontPaths()...)
	case "darwin":
		fontPaths = append(fontPaths, getMacFontPaths()...)
	case "linux":
		fontPaths = append(fontPaths, getLinuxFontPaths()...)
	default:
		return nil
	}

	sortByLength(fontPaths)
	return fontPaths
}

func getWindowsFontPaths() []string {
	var fontPaths []string

	// System Fonts
	systemFontsDir := "C:\\Windows\\Fonts"
	getFontsRecursively(systemFontsDir, &fontPaths)

	// User Fonts
	userFontsDir := os.ExpandEnv("%APPDATA%\\Microsoft\\Windows\\Fonts")
	getFontsRecursively(userFontsDir, &fontPaths)

	return fontPaths
}

func getMacFontPaths() []string {
	var fontPaths []string

	// System Fonts
	systemFontsDirs := []string{"/System/Library/Fonts", "/Library/Fonts"}
	for _, dir := range systemFontsDirs {
		getFontsRecursively(dir, &fontPaths)
	}

	// User Fonts
	userFontsDir := filepath.Join(os.Getenv("HOME"), "Library/Fonts")
	getFontsRecursively(userFontsDir, &fontPaths)
	return fontPaths
}

func getLinuxFontPaths() []string {
	var fontPaths []string

	// System Fonts
	systemFontsDirs := []string{"/usr/share/fonts", "/usr/local/share/fonts"}
	for _, dir := range systemFontsDirs {
		getFontsRecursively(dir, &fontPaths)
	}

	// User Fonts
	userFontsDir := filepath.Join(os.Getenv("HOME"), ".fonts")
	getFontsRecursively(userFontsDir, &fontPaths)

	return fontPaths
}

func getFontsRecursively(dir string, fontPaths *[]string) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return
	}

	for _, file := range files {
		path := filepath.Join(dir, file.Name())
		if file.IsDir() {
			getFontsRecursively(path, fontPaths)
		} else if strings.HasSuffix(strings.ToLower(file.Name()), ".ttf") {
			*fontPaths = append(*fontPaths, path)
		}
	}
}

func Render(t *element.Text) (*image.RGBA, int) {
	if t.LineHeight == 0 {
		t.LineHeight = t.EM + 3
	}

	width := MeasureText(t, t.Text+" ")

	// Use fully transparent color for the background
	img := image.NewRGBA(image.Rect(0, 0, width, t.LineHeight))

	// fmt.Println(t.Width, t.LineHeight, (len(lines)))

	r, g, b, a := t.Color.RGBA()

	draw.Draw(img, img.Bounds(), &image.Uniform{color.RGBA{uint8(r), uint8(g), uint8(b), uint8(0)}}, image.Point{}, draw.Over)
	// fmt.Println(int(t.Font.Metrics().Ascent))
	dot := fixed.Point26_6{X: fixed.I(0), Y: (fixed.I(t.LineHeight+(t.EM/2)) / 2)}

	dr := &font.Drawer{
		Dst:  img,
		Src:  &image.Uniform{color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)}},
		Face: *t.Font,
		Dot:  dot,
	}

	drawn := drawString(*t, dr, t.Text, width, img)

	return drawn, width
}

func drawString(t element.Text, dr *font.Drawer, v string, lineWidth int, img *image.RGBA) *image.RGBA {
	underlinePosition := dr.Dot
	for _, ch := range v {
		if ch == ' ' {
			// Handle spaces separately, add word spacing
			dr.Dot.X += fixed.I(t.WordSpacing)
		} else {
			dr.DrawString(string(ch))
			dr.Dot.X += fixed.I(t.LetterSpacing)
		}
	}
	if t.Underlined || t.Overlined || t.LineThrough {

		underlinePosition.X = 0
		baseLineY := underlinePosition.Y
		fnt := *t.Font
		descent := fnt.Metrics().Descent
		if t.Underlined {
			underlinePosition.Y = baseLineY + descent
			underlinePosition.Y = (underlinePosition.Y / 100) * 97
			drawLine(img, underlinePosition, fixed.Int26_6(lineWidth), t.DecorationThickness, t.DecorationColor)
		}
		if t.LineThrough {
			underlinePosition.Y = baseLineY - (descent)
			drawLine(img, underlinePosition, fixed.Int26_6(lineWidth), t.DecorationThickness, t.DecorationColor)
		}
		if t.Overlined {
			underlinePosition.Y = baseLineY - descent*3
			drawLine(img, underlinePosition, fixed.Int26_6(lineWidth), t.DecorationThickness, t.DecorationColor)
		}
	}
	return img
}

func drawLine(img draw.Image, start fixed.Point26_6, width fixed.Int26_6, thickness int, col color.Color) {
	// Bresenham's line algorithm
	x0, y0 := start.X.Round(), start.Y.Round()
	x1 := x0 + int(width)
	y1 := y0
	dx := abs(x1 - x0)
	dy := abs(y1 - y0)
	sx, sy := 1, 1

	if x0 > x1 {
		sx = -1
	}
	if y0 > y1 {
		sy = -1
	}

	err := dx - dy

	for {
		for i := 0; i < thickness; i++ {
			img.Set(x0, (y0-(thickness/2))+i, col)
		}

		if x0 == x1 && y0 == y1 {
			break
		}

		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func Min(a, b float32) float32 {
	if a < b {
		return a
	} else {
		return b
	}
}
