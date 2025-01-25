package font

import (
	adapter "grim/adapters"
	"grim/canvas"
	"grim/element"
	"grim/utils"
	"image"
	"image/color"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	cc "grim/color"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
)

type MetaData struct {
	Font                *font.Face
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
func GetFontPath(fontName string, bold string, italic bool, fs *adapter.FileSystem) string {
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
			fontPath = findFont("Georgia", bold, italic, paths)
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
	namePattern := `(?i)\b` + regexp.QuoteMeta(strings.ToLower(name)) + `\b` // Match 'name' as a word, case-insensitive
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
		matched, _ := regexp.MatchString(namePattern, fileName)
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

func LoadFont(fontName string, fontSize int, bold string, italic bool, fs *adapter.FileSystem) (font.Face, error) {
	// Use a TrueType font file for the specified font name
	fontFile := GetFontPath(fontName, bold, italic, fs)

	// Read the font file
	fontData, err := fs.ReadFile(fontFile)
	if err != nil {
		return nil, err
	}

	// Parse the TrueType font data
	fnt, err := truetype.Parse(fontData)
	if err != nil {
		return nil, err
	}

	options := truetype.Options{
		Size:    (float64(fontSize) * 72) / 96,
		DPI:     96,
		Hinting: font.HintingNone,
	}

	// Create a new font face with the specified size
	return truetype.NewFace(fnt, &options), nil
}

func MeasureText(t *MetaData, text string) (int, int) {
	ctx := canvas.NewCanvas(0, 0)
	ctx.Context.SetFontFace(*t.Font)
	w, h := ctx.MeasureText(text)
	return int(w), int(h)
}

func MeasureSpace(t *MetaData) (int, int) {
	ctx := canvas.NewCanvas(0, 0)
	ctx.Context.SetFontFace(*t.Font)
	w, h := ctx.MeasureText(" ")
	return int(w), int(h)
}

func Key(text *MetaData) string {
	key := text.Text + utils.RGBAtoString(text.Color) + utils.RGBAtoString(text.DecorationColor) + text.Align + text.WordBreak + strconv.Itoa(text.WordSpacing) + strconv.Itoa(text.LetterSpacing) + text.WhiteSpace + strconv.Itoa(text.DecorationThickness) + strconv.Itoa(text.EM)
	key += strconv.FormatBool(text.Overlined) + strconv.FormatBool(text.Underlined) + strconv.FormatBool(text.LineThrough)
	return key
}

func GetMetaData(n *element.Node, state *map[string]element.State, font *font.Face) *MetaData {
	s := *state
	self := s[n.Properties.Id]
	parent := s[n.Parent.Properties.Id]

	// self.Textures = []string{}

	text := MetaData{}
	text.Font = font
	letterSpacing := utils.ConvertToPixels(n.CStyle["letter-spacing"], self.EM, parent.Width)
	wordSpacing := utils.ConvertToPixels(n.CStyle["word-spacing"], self.EM, parent.Width)
	lineHeight := utils.ConvertToPixels(n.CStyle["line-height"], self.EM, parent.Width)
	underlineoffset := utils.ConvertToPixels(n.CStyle["text-underline-offset"], self.EM, parent.Width)

	if lineHeight == 0 {
		lineHeight = self.EM + 3
	}

	text.LineHeight = int(lineHeight)
	text.WordSpacing = int(wordSpacing)
	text.LetterSpacing = int(letterSpacing)
	wb := " "

	if n.CStyle["word-wrap"] == "break-word" {
		wb = ""
	}

	if n.CStyle["text-wrap"] == "wrap" || n.CStyle["text-wrap"] == "balance" {
		wb = ""
	}

	var dt float32

	if n.CStyle["text-decoration-thickness"] == "auto" || n.CStyle["text-decoration-thickness"] == "" {
		dt = self.EM / 7
	} else {
		dt = utils.ConvertToPixels(n.CStyle["text-decoration-thickness"], self.EM, parent.Width)
	}

	col := cc.Parse(n.CStyle, "font")

	if n.CStyle["text-decoration-color"] == "" {
		n.CStyle["text-decoration-color"] = n.CStyle["color"]
	}

	text.Color = col
	text.DecorationColor = cc.Parse(n.CStyle, "decoration")
	text.Align = n.CStyle["text-align"]
	text.WordBreak = wb
	text.WordSpacing = int(wordSpacing)
	text.LetterSpacing = int(letterSpacing)
	text.WhiteSpace = n.CStyle["white-space"]
	text.DecorationThickness = int(dt)
	text.Overlined = n.CStyle["text-decoration"] == "overline"
	text.Underlined = n.CStyle["text-decoration"] == "underline"
	text.LineThrough = n.CStyle["text-decoration"] == "line-through"
	text.EM = int(self.EM)
	text.Width = int(parent.Width)
	text.Text = n.InnerText
	text.UnderlineOffset = int(underlineoffset)

	if n.CStyle["text-underline-offset"] == "" {
		text.UnderlineOffset = 2
	}

	if n.CStyle["word-spacing"] == "" {
		// !ISSUE: is word spacing actually impleamented
		text.WordSpacing, _ = MeasureSpace(&text)
	}
	return &text
}

func Render(text *MetaData) (*image.RGBA, int) {
	if text.LineHeight == 0 {
		text.LineHeight = text.EM + 3
	}

	width, _ := MeasureText(text, text.Text+" ")

	ctx := canvas.NewCanvas(width, text.LineHeight)
	r, g, b, a := text.Color.RGBA()

	ctx.SetFillStyle(uint8(r), uint8(g), uint8(b), uint8(a))
	ctx.Context.SetFontFace(*text.Font)
	ctx.Context.DrawStringAnchored(text.Text, 0, float64(text.LineHeight)/2, 0, 0.3)

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
	return ctx.RGBA, width
}
