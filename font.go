package grim

import (
	"errors"
	"golang.org/x/image/math/fixed"
	"grim/canvas"
	"image"
	"image/color"
	"path/filepath"
	"strconv"
	"strings"

	cc "grim/color"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
)

type MetaData struct {
	Font                *truetype.Font
	FontFamily          string
	Color               color.RGBA
	Text                string
	Underlined          bool
	Overlined           bool
	LineThrough         bool
	DecorationColor     color.RGBA
	DecorationThickness int
	Align               string
	Indent              int // very low priority
	LetterSpacing       int
	LineHeight          int
	WordSpacing         int
	WhiteSpace          string
	Shadows             []Shadow // need
	Width               int
	WordBreak           string
	EM                  int
	X                   int
	UnderlineOffset     int
}

type Shadow struct {
	X     int
	Y     int
	Blur  int
	Color color.RGBA
}

// LoadSystemFont loads a font from the system fonts directory or loads a specific font by name
func GetFontPath(fontName string, bold string, italic bool, fs *FileSystem) string {
	if len(fontName) == 0 {
		fontName = "serif"
	}

	fonts := strings.Split(fontName, ",")
	paths := fs.Paths
	for _, font := range fonts {
		font = strings.TrimSpace(font)
		var fontPath string

		// Check special font families only if it's the first font in the list
		switch font {
		case "sans-serif":
			fontPath = findFont("Arial", bold, italic, paths)
		case "monospace":
			fontPath = findFont("Andale Mono", bold, italic, paths)
		case "serif":
			fontPath = findFont("Times New Roman", bold, italic, paths)
		default:
			fontPath = findFont(font, bold, italic, paths)
		}

		if fontPath != "" {
			return fontPath
		}

	}
	// Default to serif if none of the specified fonts are found
	return findFont("Georgia", bold, italic, paths)
}

func findFont(name string, bold string, italic bool, paths []string) string {
	wns := []string{"thin",
		"extralight",
		"light",
		"medium",
		"semibold",
		"bold",
		"extrabold",
		"black"}
	matches := []string{}
	for _, v := range paths {
		fileName := filepath.Base(strings.ToLower(v))
		matched := strings.Contains(strings.ToLower(fileName), strings.ToLower(name))
		if matched {

			if italic {
				if strings.Contains(fileName, "italic") {
					matches = append(matches, strings.ToLower(v))
				}
			} else {
				if !strings.Contains(fileName, "italic") {
					matches = append(matches, strings.ToLower(v))
				}
			}
		}
	}

	for _, v := range matches {
		weightName := GetWeightName(bold)
		if bold == "" {
			doesContain := false
			for _, wn := range wns {
				if strings.Contains(v, wn) {
					doesContain = true
				}
			}
			if doesContain {
				continue
			}
		}
		if strings.Contains(v, weightName) {
			return v
		}
	}
	if len(matches) > 0 {
		return matches[0]
	} else {
		return ""
	}
}

func GetWeightName(weight string) string {
	switch weight {
	case "100":
		return "thin"
	case "200":
		return "extralight"
	case "300":
		return "light"
	case "400":
		return ""
	case "500":
		return "medium"
	case "600":
		return "semibold"
	case "700":
		return "bold"
	case "800":
		return "extrabold"
	case "900":
		return "black"
	default:
		return weight
	}
}

func LoadFont(fontName string, fontSize int, bold string, italic bool, fs *FileSystem) (*truetype.Font, error) {
	// Use a TrueType font file for the specified font name
	fontFile := GetFontPath(fontName, bold, italic, fs)
	// Read the font file
	fontData, err := fs.ReadFile(fontFile)
	if err != nil {
		return nil, errors.New("Font file not found!")
	}

	// Parse the TrueType font data
	fnt, err := truetype.Parse(fontData)
	if err != nil {
		return nil, errors.New("Unable to parse font data!\n")
	}
	return fnt, nil
}

func MeasureText(t *MetaData, text string) int {
	// Create font options
	options := truetype.Options{
		Size:    (float64(t.EM) * 72) / 96,
		DPI:     96,
		Hinting: font.HintingNone,
	}

	// Calculate scale factor
	scale := (options.Size * options.DPI) / (72.0 * float64(t.Font.FUnitsPerEm()))

	var width float64
	for _, r := range text {
		idx := t.Font.Index(r)
		hMetric := t.Font.HMetric(fixed.Int26_6(t.Font.FUnitsPerEm()), idx)
		width += float64(hMetric.AdvanceWidth) * scale
	}
	return int(width)
}

func MeasureSpace(t *MetaData) int {
	return MeasureText(t, " ")
}

func FontKey(text *MetaData) string {
	key := text.Text + RGBAtoString(text.Color) + RGBAtoString(text.DecorationColor) + text.Align + text.WordBreak + strconv.Itoa(text.WordSpacing) + strconv.Itoa(text.LetterSpacing) + text.WhiteSpace + strconv.Itoa(text.DecorationThickness) + strconv.Itoa(text.EM)
	key += strconv.FormatBool(text.Overlined) + strconv.FormatBool(text.Underlined) + strconv.FormatBool(text.LineThrough) + text.FontFamily
	return key
}

func GetMetaData(n *Node, style map[string]string, state *map[string]State, font *truetype.Font) *MetaData {
	s := *state
	self := s[n.Properties.Id]
	parent := s[n.Parent().Properties.Id]

	// !DEVMAN: In some cases like a span the width will be unset. In that case
	// + find the closest parent that has a width greater than 0
	if parent.Width == 0 {
		ancestors := strings.Split(n.Properties.Id, ":")

		var parentId string
		// Should skip the current element and the ROOT
		for i := len(ancestors) - 2; i > 0; i-- {
			parentId = strings.Join(ancestors[0:i], ":")
			if s[parentId].Width > 0 {
				parent = s[parentId]
				break
			}
		}
	}

	// self.Textures = []string{}

	text := MetaData{}
	text.Font = font
	text.FontFamily = style["font-family"]
	letterSpacing := ConvertToPixels(style["letter-spacing"], self.EM, parent.Width)
	wordSpacing := ConvertToPixels(style["word-spacing"], self.EM, parent.Width)
	lineHeight := ConvertToPixels(style["line-height"], self.EM, parent.Width)
	underlineoffset := ConvertToPixels(style["text-underline-offset"], self.EM, parent.Width)

	if lineHeight == 0 {
		lineHeight = self.EM + 3
	}

	text.LineHeight = int(lineHeight)
	text.WordSpacing = int(wordSpacing)
	text.LetterSpacing = int(letterSpacing)
	wb := " "

	if style["word-wrap"] == "break-word" {
		wb = ""
	}

	if style["text-wrap"] == "wrap" || style["text-wrap"] == "balance" {
		wb = ""
	}

	var dt float32

	if style["text-decoration-thickness"] == "auto" || style["text-decoration-thickness"] == "" {
		dt = self.EM / 7
	} else {
		dt = ConvertToPixels(style["text-decoration-thickness"], self.EM, parent.Width)
	}

	col, err := cc.ParseRGBA(style["color"])

	if err != nil {
		col = color.RGBA{0, 0, 0, 255}
	}

	if style["text-decoration-color"] == "" {
		style["text-decoration-color"] = style["color"]
	}

	text.Color = col
	text.DecorationColor, _ = cc.ParseRGBA(style["text-decoration-color"])
	text.Align = style["text-align"]
	text.WordBreak = wb
	text.WordSpacing = int(wordSpacing)
	text.LetterSpacing = int(letterSpacing)
	text.WhiteSpace = style["white-space"]
	text.DecorationThickness = int(dt)
	text.Overlined = style["text-decoration"] == "overline"
	text.Underlined = style["text-decoration"] == "underline"
	text.LineThrough = style["text-decoration"] == "line-through"
	text.EM = int(self.EM)
	text.Width = int(parent.Width)
	text.Text = n.InnerText()
	text.UnderlineOffset = int(underlineoffset)

	if style["text-underline-offset"] == "" {
		text.UnderlineOffset = 2
	}

	// if style["word-spacing"] == "" {
	// 	// !ISSUE: is word spacing actually impleamented
	// 	text.WordSpacing = MeasureSpace(&text)
	// }
	return &text
}

func RenderFont(text *MetaData) (image.Image, int) {
	if text.LineHeight == 0 {
		text.LineHeight = text.EM + 3
	}

	options := truetype.Options{
		Size:    (float64(text.EM) * 72) / 96,
		DPI:     96,
		Hinting: font.HintingNone,
	}

	// Create a new font face with the specified size
	font := truetype.NewFace(text.Font, &options)

	width := MeasureText(text, text.Text+" ")

	ctx := canvas.NewCanvas(width, text.LineHeight)
	r, g, b, a := text.Color.RGBA()

	ctx.SetFillStyle(uint8(r), uint8(g), uint8(b), uint8(a))
	ctx.Context.SetFontFace(font)
	ctx.Context.DrawStringAnchored(text.Text, 0, float64(text.LineHeight)/2, 0, 0.3)
	font.Close()
	if text.Underlined || text.Overlined || text.LineThrough {
		ctx.SetLineWidth(float64(text.DecorationThickness))
		r, g, b, a = text.DecorationColor.RGBA()
		ctx.SetStrokeStyle(uint8(r), uint8(g), uint8(b), uint8(a))
		ctx.BeginPath()
		var y float64
		if text.Underlined {
			y = (float64(text.LineHeight) / 2) + (float64(text.EM) / 2.5) + float64(text.UnderlineOffset)
		}
		if text.LineThrough {
			y = (float64(text.LineHeight) / 2)
		}
		if text.Overlined {
			y = (float64(text.LineHeight) / 2) - (float64(text.EM) / 2) - (float64(text.DecorationThickness) / 2)
		}
		ctx.MoveTo(0, y)
		ctx.LineTo(float64(width), y)
		ctx.Stroke()
	}
	return ctx.Context.Image(), width
}
