package flexprep

import (
	"fmt"
	"grim/cstyle"
	"grim/element"
	"grim/lgc"
	"strconv"
	"strings"
)

func Init() cstyle.Transformer {
	return cstyle.Transformer{
		Selector: func(n *element.Node, c *cstyle.CSS) bool {
			return n.GetStyle("flex") != ""
		},
		Handler: func(n *element.Node, c *cstyle.CSS) *element.Node {
			flex, err := parseFlex(n.GetStyle("flex"))

			if err != nil {
				panic(lgc.Error(
					lgc.Desc(
						lgc.Red("Error Parsing flex values"),
						lgc.CrossMark("Input: "+lgc.String(n.GetStyle("flex"))),
						err,
					),
				))
			}

			n.SetStyle("flex-basis", flex.FlexBasis)
			n.SetStyle("flex-grow", flex.FlexGrow)
			n.SetStyle("flex-shrink", flex.FlexShrink)

			return n
		},
	}
}

type FlexProperties struct {
	FlexGrow   string
	FlexShrink string
	FlexBasis  string
}

func parseFlex(flex string) (FlexProperties, error) {
	parts := strings.Fields(flex)
	prop := FlexProperties{
		FlexGrow:   "1",  // default value
		FlexShrink: "1",  // default value
		FlexBasis:  "0%", // default value
	}

	switch len(parts) {
	case 1:
		if strings.HasSuffix(parts[0], "%") || strings.HasSuffix(parts[0], "px") || strings.HasSuffix(parts[0], "em") {
			prop.FlexBasis = parts[0]
		} else if _, err := strconv.ParseFloat(parts[0], 64); err == nil {
			prop.FlexGrow = parts[0]
			prop.FlexShrink = "1"
			prop.FlexBasis = "0%"
		} else {
			return prop, fmt.Errorf("invalid flex value: %s", parts[0])
		}
	case 2:
		prop.FlexGrow = parts[0]
		prop.FlexShrink = parts[1]
		prop.FlexBasis = "0%"
	case 3:
		prop.FlexGrow = parts[0]
		prop.FlexShrink = parts[1]
		prop.FlexBasis = parts[2]
	default:
		return prop, fmt.Errorf("invalid number of values for flex property")
	}

	return prop, nil
}
