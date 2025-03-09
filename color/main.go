package color

import (
	"fmt"
	ic "image/color"
	"strconv"
	"strings"
)

// ParseRGBA parses a CSS color string and returns an RGBA color
func ParseRGBA(color string) (ic.RGBA, error) {
	color = strings.TrimSpace(strings.ToLower(color))

	// Named color
	if namedColor, ok := namedColors[color]; ok {
		return namedColor, nil
	}

	// Hex color format: #RRGGBB or #RRGGBBAA
	if strings.HasPrefix(color, "#") {
		hexValue := strings.TrimPrefix(color, "#")

		// Check if it's a three-letter hex code and expand it to six-letter hex code
		if len(hexValue) == 3 {
			hexValue = fmt.Sprintf("%c%c%c%c%c%c",
				hexValue[0], hexValue[0],
				hexValue[1], hexValue[1],
				hexValue[2], hexValue[2])
		}

		rgb, err := strconv.ParseUint(hexValue, 16, 32)
		if err != nil {
			return ic.RGBA{}, fmt.Errorf("Invalid hex code")
		}

		// Check if it's #RRGGBB or #RRGGBBAA
		alpha := uint8(255)
		if len(hexValue) == 8 {
			alpha = uint8(rgb >> 24)
		}

		return ic.RGBA{uint8(rgb >> 16), uint8((rgb >> 8) & 0xFF), uint8(rgb & 0xFF), alpha}, nil
	}

	// RGB or RGBA color format: rgb(255, 0, 0) or rgba(255, 0, 0, 0.5)
	if strings.HasPrefix(color, "rgb(") && strings.HasSuffix(color, ")") {
		return parseRGB(color)
	} else if strings.HasPrefix(color, "rgba(") && strings.HasSuffix(color, ")") {
		return parseRGBA(color)
	}

	// HSL or HSLA color format: hsl(0, 100%, 50%) or hsla(0, 100%, 50%, 0.5)
	if strings.HasPrefix(color, "hsl(") && strings.HasSuffix(color, ")") {
		return parseHSL(color)
	} else if strings.HasPrefix(color, "hsla(") && strings.HasSuffix(color, ")") {
		return parseHSLA(color)
	}

	return ic.RGBA{}, fmt.Errorf("Unable to parse color")
}

func parseRGB(rgb string) (ic.RGBA, error) {
	rgbValues := strings.TrimSuffix(strings.TrimPrefix(rgb, "rgb("), ")")
	rgbParts := strings.Split(rgbValues, ",")

	if len(rgbParts) != 3 {
		return ic.RGBA{}, fmt.Errorf("Not enough values to parse RGB")
	}

	r, err := strconv.Atoi(strings.TrimSpace(rgbParts[0]))
	if err != nil {
		return ic.RGBA{}, err
	}

	g, err := strconv.Atoi(strings.TrimSpace(rgbParts[1]))
	if err != nil {
		return ic.RGBA{}, err
	}

	b, err := strconv.Atoi(strings.TrimSpace(rgbParts[2]))
	if err != nil {
		return ic.RGBA{}, err
	}

	return ic.RGBA{uint8(r), uint8(g), uint8(b), 255}, nil
}

func parseRGBA(rgba string) (ic.RGBA, error) {
	rgbaValues := strings.TrimSuffix(strings.TrimPrefix(rgba, "rgba("), ")")
	rgbaParts := strings.Split(rgbaValues, ",")

	if len(rgbaParts) != 4 {
		return ic.RGBA{}, fmt.Errorf("Not enough values to parse RGBA")
	}

	r, err := strconv.Atoi(strings.TrimSpace(rgbaParts[0]))
	if err != nil {
		return ic.RGBA{}, err
	}

	g, err := strconv.Atoi(strings.TrimSpace(rgbaParts[1]))
	if err != nil {
		return ic.RGBA{}, err
	}

	b, err := strconv.Atoi(strings.TrimSpace(rgbaParts[2]))
	if err != nil {
		return ic.RGBA{}, err
	}

	alpha, err := strconv.ParseFloat(strings.TrimSpace(rgbaParts[3]), 64)
	if err != nil {
		return ic.RGBA{}, err
	}

	return ic.RGBA{uint8(r), uint8(g), uint8(b), uint8(alpha * 255)}, nil
}

func parseHSL(hsl string) (ic.RGBA, error) {
	hslValues := strings.TrimSuffix(strings.TrimPrefix(hsl, "hsl("), ")")
	hslParts := strings.Split(hslValues, ",")

	if len(hslParts) != 3 {
		return ic.RGBA{}, fmt.Errorf("Not enough values to parse HSL")
	}

	h, err := strconv.Atoi(strings.TrimSpace(hslParts[0]))
	if err != nil {
		return ic.RGBA{}, err
	}

	s, err := strconv.Atoi(strings.TrimSpace(strings.TrimSuffix(hslParts[1], "%")))
	if err != nil {
		return ic.RGBA{}, err
	}

	l, err := strconv.Atoi(strings.TrimSpace(strings.TrimSuffix(hslParts[2], "%")))
	if err != nil {
		return ic.RGBA{}, err
	}

	return hslToRGB(uint16(h), float64(s)/100, float64(l)/100)
}

func parseHSLA(hsla string) (ic.RGBA, error) {
	hslaValues := strings.TrimSuffix(strings.TrimPrefix(hsla, "hsla("), ")")
	hslaParts := strings.Split(hslaValues, ",")

	if len(hslaParts) != 4 {
		return ic.RGBA{}, fmt.Errorf("Not enough values to parse HSLA")
	}

	h, err := strconv.Atoi(strings.TrimSpace(hslaParts[0]))
	if err != nil {
		return ic.RGBA{}, err
	}

	s, err := strconv.Atoi(strings.TrimSpace(strings.TrimSuffix(hslaParts[1], "%")))
	if err != nil {
		return ic.RGBA{}, err
	}

	l, err := strconv.Atoi(strings.TrimSpace(strings.TrimSuffix(hslaParts[2], "%")))
	if err != nil {
		return ic.RGBA{}, err
	}

	alpha, err := strconv.ParseFloat(strings.TrimSpace(hslaParts[3]), 64)
	if err != nil {
		return ic.RGBA{}, err
	}

	return hslToRGB(uint16(h), float64(s)/100, float64(l)/100, alpha)
}

func hslToRGB(hue uint16, saturation, lightness float64, alpha ...float64) (ic.RGBA, error) {
	var r, g, b float64

	if saturation == 0 {
		r = lightness
		g = lightness
		b = lightness
	} else {
		var q float64
		if lightness < 0.5 {
			q = lightness * (1 + saturation)
		} else {
			q = lightness + saturation - (lightness * saturation)
		}

		p := 2*lightness - q

		h := float64(hue) / 360

		// Convert hue to RGB
		var tc [3]float64
		tc[0] = h + 1/3
		tc[1] = h
		tc[2] = h - 1/3

		for i := 0; i < 3; i++ {
			if tc[i] < 0 {
				tc[i] += 1
			} else if tc[i] > 1 {
				tc[i] -= 1
			}

			if tc[i] < 1/6 {
				tc[i] = p + (q-p)*6*tc[i]
			} else if tc[i] < 1/2 {
				tc[i] = q
			} else if tc[i] < 2/3 {
				tc[i] = p + (q-p)*6*(2/3-tc[i])
			} else {
				tc[i] = p
			}
		}

		r, g, b = tc[0], tc[1], tc[2]
	}

	// Scale to 0-255
	r *= 255
	g *= 255
	b *= 255

	var alphaValue uint8 = 255
	if len(alpha) > 0 {
		alphaValue = uint8(alpha[0] * 255)
	}

	return ic.RGBA{uint8(r), uint8(g), uint8(b), alphaValue}, nil
}

var namedColors = map[string]ic.RGBA{
	"aliceblue":            {240, 248, 255, 255},
	"antiquewhite":         {250, 235, 215, 255},
	"aqua":                 {0, 255, 255, 255},
	"aquamarine":           {127, 255, 212, 255},
	"azure":                {240, 255, 255, 255},
	"beige":                {245, 245, 220, 255},
	"bisque":               {255, 228, 196, 255},
	"black":                {0, 0, 0, 255},
	"blanchedalmond":       {255, 235, 205, 255},
	"blue":                 {0, 0, 255, 255},
	"blueviolet":           {138, 43, 226, 255},
	"brown":                {165, 42, 42, 255},
	"burlywood":            {222, 184, 135, 255},
	"cadetblue":            {95, 158, 160, 255},
	"chartreuse":           {127, 255, 0, 255},
	"chocolate":            {210, 105, 30, 255},
	"coral":                {255, 127, 80, 255},
	"cornflowerblue":       {100, 149, 237, 255},
	"cornsilk":             {255, 248, 220, 255},
	"crimson":              {220, 20, 60, 255},
	"cyan":                 {0, 255, 255, 255},
	"darkblue":             {0, 0, 139, 255},
	"darkcyan":             {0, 139, 139, 255},
	"darkgoldenrod":        {184, 134, 11, 255},
	"darkgray":             {169, 169, 169, 255},
	"darkgreen":            {0, 100, 0, 255},
	"darkkhaki":            {189, 183, 107, 255},
	"darkmagenta":          {139, 0, 139, 255},
	"darkolivegreen":       {85, 107, 47, 255},
	"darkorange":           {255, 140, 0, 255},
	"darkorchid":           {153, 50, 204, 255},
	"darkred":              {139, 0, 0, 255},
	"darksalmon":           {233, 150, 122, 255},
	"darkseagreen":         {143, 188, 143, 255},
	"darkslateblue":        {72, 61, 139, 255},
	"darkslategray":        {47, 79, 79, 255},
	"darkturquoise":        {0, 206, 209, 255},
	"darkviolet":           {148, 0, 211, 255},
	"deeppink":             {255, 20, 147, 255},
	"deepskyblue":          {0, 191, 255, 255},
	"dimgray":              {105, 105, 105, 255},
	"dodgerblue":           {30, 144, 255, 255},
	"firebrick":            {178, 34, 34, 255},
	"floralwhite":          {255, 250, 240, 255},
	"forestgreen":          {34, 139, 34, 255},
	"fuchsia":              {255, 0, 255, 255},
	"gainsboro":            {220, 220, 220, 255},
	"ghostwhite":           {248, 248, 255, 255},
	"gold":                 {255, 215, 0, 255},
	"goldenrod":            {218, 165, 32, 255},
	"gray":                 {128, 128, 128, 255},
	"green":                {0, 128, 0, 255},
	"greenyellow":          {173, 255, 47, 255},
	"honeydew":             {240, 255, 240, 255},
	"hotpink":              {255, 105, 180, 255},
	"indianred":            {205, 92, 92, 255},
	"indigo":               {75, 0, 130, 255},
	"ivory":                {255, 255, 240, 255},
	"khaki":                {240, 230, 140, 255},
	"lavender":             {230, 230, 250, 255},
	"lavenderblush":        {255, 240, 245, 255},
	"lawngreen":            {124, 252, 0, 255},
	"lemonchiffon":         {255, 250, 205, 255},
	"lightblue":            {173, 216, 230, 255},
	"lightcoral":           {240, 128, 128, 255},
	"lightcyan":            {224, 255, 255, 255},
	"lightgoldenrodyellow": {250, 250, 210, 255},
	"lightgreen":           {144, 238, 144, 255},
	"lightgrey":            {211, 211, 211, 255},
	"lightpink":            {255, 182, 193, 255},
	"lightsalmon":          {255, 160, 122, 255},
	"lightseagreen":        {32, 178, 170, 255},
	"lightskyblue":         {135, 206, 250, 255},
	"lightslategray":       {119, 136, 153, 255},
	"lightsteelblue":       {176, 196, 222, 255},
	"lightyellow":          {255, 255, 224, 255},
	"lime":                 {0, 255, 0, 255},
	"limegreen":            {50, 205, 50, 255},
	"linen":                {250, 240, 230, 255},
	"magenta":              {255, 0, 255, 255},
	"maroon":               {128, 0, 0, 255},
	"mediumaquamarine":     {102, 205, 170, 255},
	"mediumblue":           {0, 0, 205, 255},
	"mediumorchid":         {186, 85, 211, 255},
	"mediumpurple":         {147, 112, 219, 255},
	"mediumseagreen":       {60, 179, 113, 255},
	"mediumslateblue":      {123, 104, 238, 255},
	"mediumspringgreen":    {0, 250, 154, 255},
	"mediumturquoise":      {72, 209, 204, 255},
	"mediumvioletred":      {199, 21, 133, 255},
	"midnightblue":         {25, 25, 112, 255},
	"mintcream":            {245, 255, 250, 255},
	"mistyrose":            {255, 228, 225, 255},
	"moccasin":             {255, 228, 181, 255},
	"navajowhite":          {255, 222, 173, 255},
	"navy":                 {0, 0, 128, 255},
	"oldlace":              {253, 245, 230, 255},
	"olive":                {128, 128, 0, 255},
	"olivedrab":            {107, 142, 35, 255},
	"orange":               {255, 165, 0, 255},
	"orangered":            {255, 69, 0, 255},
	"orchid":               {218, 112, 214, 255},
	"palegoldenrod":        {238, 232, 170, 255},
	"palegreen":            {152, 251, 152, 255},
	"paleturquoise":        {175, 238, 238, 255},
	"palevioletred":        {219, 112, 147, 255},
	"papayawhip":           {255, 239, 213, 255},
	"peachpuff":            {255, 218, 185, 255},
	"peru":                 {205, 133, 63, 255},
	"pink":                 {255, 192, 203, 255},
	"plum":                 {221, 160, 221, 255},
	"powderblue":           {176, 224, 230, 255},
	"purple":               {128, 0, 128, 255},
	"red":                  {255, 0, 0, 255},
	"rosybrown":            {188, 143, 143, 255},
	"royalblue":            {65, 105, 225, 255},
	"saddlebrown":          {139, 69, 19, 255},
	"salmon":               {250, 128, 114, 255},
	"sandybrown":           {244, 164, 96, 255},
	"seagreen":             {46, 139, 87, 255},
	"seashell":             {255, 245, 238, 255},
	"sienna":               {160, 82, 45, 255},
	"silver":               {192, 192, 192, 255},
	"skyblue":              {135, 206, 235, 255},
	"slateblue":            {106, 90, 205, 255},
	"slategray":            {112, 128, 144, 255},
	"snow":                 {255, 250, 250, 255},
	"springgreen":          {0, 255, 127, 255},
	"steelblue":            {70, 130, 180, 255},
	"tan":                  {210, 180, 140, 255},
	"teal":                 {0, 128, 128, 255},
	"thistle":              {216, 191, 216, 255},
	"tomato":               {255, 99, 71, 255},
	"turquoise":            {64, 224, 208, 255},
	"violet":               {238, 130, 238, 255},
	"wheat":                {245, 222, 179, 255},
	"white":                {255, 255, 255, 255},
	"whitesmoke":           {245, 245, 245, 255},
	"yellow":               {255, 255, 0, 255},
	"yellowgreen":          {154, 205, 50, 255},
}

func Parse(styles map[string]string, t string) (ic.RGBA, error) {
	if t == "font" {
		return ParseRGBA(styles["color"])
	} else if t == "decoration" {
		return ParseRGBA(styles["text-decoration-color"])
	} else if t == "background" {
		backgroundColor := styles["background-color"]
		// Parse the background color and return the result
		return ParseRGBA(backgroundColor)
	} else {
		return ic.RGBA{255, 255, 255, 0}, fmt.Errorf("Color not found")
	}
}

func Color(color string) (ic.RGBA, error) {
	if color == "" {
		color = "rgba(0,0,0,1)"
	}

	// Parse the background color and return the result
	return ParseRGBA(color)
}
