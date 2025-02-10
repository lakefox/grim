package lgc

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
)

// Define log levels using bitmask flags
const (
	LogDebug = 1 << iota
	LogInfo
	LogWarn
	LogError
	LogFatal
)

// Define colors
const (
	ColorReset  = "\033[0m"
	ColorGray   = "\033[90m"
	ColorBlue   = "\033[34m"
	ColorYellow = "\033[33m"
	ColorRed    = "\033[31m"
	ColorPurple = "\033[35m"
	ColorGreen  = "\033[32m"
)

// Active log levels (default: all enabled)
var logLevels = LogDebug | LogInfo | LogWarn | LogError | LogFatal

// Set specific log levels
func SetLogLevels(levels int) {
	logLevels = levels
}

// Enable additional log levels
func EnableLogLevels(levels int) {
	logLevels |= levels
}

// Disable specific log levels
func DisableLogLevels(levels int) {
	logLevels &^= levels
}

// Helper function to get caller info
func getCallerInfo(color string) string {
	_, file, line, ok := runtime.Caller(2) // Get caller's information (2 levels up)
	if !ok {
		return "[unknown]"
	}
	return fmt.Sprintf("\n"+color+"┃"+ColorReset+" %s:%d\n"+color, file, line)
}

// Helper function to format log messages
func printLogs(color string, log ...interface{}) string {
	var sb strings.Builder
	for _, v := range log {
		sb.WriteString(color + "┃" + ColorReset + "\t " + fmt.Sprint(v) + "\n")
	}
	sb.WriteString("\n")
	return sb.String()
}

func print(colorFunc func(string) string, log ...func(func(string) string) string) string {
	var sb strings.Builder
	for _, v := range log {
		sb.WriteString(v(colorFunc) + "\n")
	}
	sb.WriteString("\n")
	return sb.String()
}

// **Composable Message Modifiers**
func Message(msg string) string       { return msg }
func String(value string) string      { return fmt.Sprintf(Yellow("\"")+Gray("%s")+Yellow("\""), value) }
func Number(value interface{}) string { return fmt.Sprintf(Gray("%v"), value) }
func Slice(value interface{}) string {
	str := fmt.Sprintf("%v", value)
	str = str[1 : len(str)-1]
	return "[" + fmt.Sprintf(Gray("%v"), str) + "]"
}
func Bool(value bool) string {
	v := Red("false")
	if value {
		v = Green("true")
	}
	return fmt.Sprintf("%s", v)
}
func Err(value error) string {
	if value != nil {
		return value.Error()
	} else {
		return "nil"
	}
}

// **Inline Functions**
func Bullet(value string) string      { return "• " + value }
func Star(value string) string        { return "★ " + value }
func CheckMark(value string) string   { return "✔ " + value }
func CrossMark(value string) string   { return "✖ " + value }
func WarningMark(value string) string { return "⚠ " + value }
func InfoMark(value string) string    { return "ℹ " + value }

// **Section Functions**
func Caller(values ...interface{}) func(func(string) string) string {
	return (func(colorFunc func(string) string) string {
		pc, _, _, ok := runtime.Caller(3) // Get caller's information (2 levels up)
		if !ok {
			return "[unknown]"
		}
		fn := runtime.FuncForPC(pc).Name() // Get function name

		var sb strings.Builder

		sb.WriteString(colorFunc("┃ ") + Purple("Function called:"))

		sb.WriteString(colorFunc("\n┃ ") + fn + "(")

		for _, v := range values {
			sb.WriteString(colorFunc("\n┃ ") + "\t" + fmt.Sprint(v))
		}

		sb.WriteString(colorFunc("\n┃ ") + ")")

		return sb.String()
	})

}
func Fix(values ...interface{}) func(func(string) string) string {
	return (func(colorFunc func(string) string) string {

		var sb strings.Builder

		sb.WriteString(colorFunc("┃ ") + Green("How to fix this issue:"))

		for _, v := range values {
			sb.WriteString(colorFunc("\n┃   ") + fmt.Sprint(v))
		}

		return sb.String()
	})

}
func Desc(values ...interface{}) func(func(string) string) string {
	return (func(colorFunc func(string) string) string {

		var sb strings.Builder

		sb.WriteString(colorFunc("┃ ") + Blue("Description:"))

		for _, v := range values {
			sb.WriteString(colorFunc("\n┃   ") + fmt.Sprint(v))
		}

		return sb.String()
	})

}

// **Color Helper Functions**
func Red(text string) string    { return ColorRed + text + ColorReset }
func Blue(text string) string   { return ColorBlue + text + ColorReset }
func Yellow(text string) string { return ColorYellow + text + ColorReset }
func Purple(text string) string { return ColorPurple + text + ColorReset }
func Gray(text string) string   { return ColorGray + text + ColorReset }
func Green(text string) string  { return ColorGreen + text + ColorReset }

// **Logging Functions**

func Info(log ...interface{}) {
	if logLevels&LogInfo == 0 {
		return
	}
	fmt.Print("\t" + Blue("Info"))
	fmt.Print(ColorGray, printLogs(ColorBlue, log...), ColorReset)
}

func Caution(log ...interface{}) {
	if logLevels&LogWarn == 0 {
		return
	}
	fmt.Print(Yellow("CAUTION")+ColorGray, getCallerInfo(ColorYellow), printLogs(ColorYellow, log...), ColorReset)
}

func Warn(log ...interface{}) {
	if logLevels&LogWarn == 0 {
		return
	}
	fmt.Print(Green("WARN")+ColorGray, getCallerInfo(ColorGreen), printLogs(ColorGreen, log...), ColorReset)
}

func Error(log ...func(func(string) string) string) error {
	if logLevels&LogError == 0 {
		return Error(log...)
	}
	return errors.New(fmt.Sprint(Red("ERROR"), getCallerInfo(ColorRed), print(Red, log...)))
}

func Fatal(log ...interface{}) {
	if logLevels&LogFatal == 0 {
		return
	}
	fmt.Print(Purple("FATAL")+ColorGray, getCallerInfo(ColorPurple), printLogs(ColorPurple, log...), ColorReset)
}
