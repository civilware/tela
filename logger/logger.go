package logger

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/deroproject/derohe/globals"
)

const (
	timestampFormat = "01/02/2006 15:04:05"

	DEBUG = "DEBUG"
	INFO  = "INFO"
	WARN  = "WARN"
	ERROR = "ERROR"
	FATAL = "FATAL"
)

type level struct {
	debug string
	info  string
	warn  string
	err   string
	fatal string
}

var tag = level{
	debug: Color.blue + DEBUG + Color.end,
	info:  Color.green + INFO + Color.end,
	warn:  Color.yellow + WARN + Color.end,
	err:   Color.red + ERROR + Color.end,
	fatal: Color.red + FATAL + Color.end,
}

// Enable or disable colors
func EnableColors(b bool) {
	if b {
		Color.black = ANSIblack
		Color.red = ANSIred
		Color.green = ANSIgreen
		Color.yellow = ANSIyellow
		Color.blue = ANSIblue
		Color.magenta = ANSImagenta
		Color.cyan = ANSIcyan
		Color.white = ANSIwhite
		Color.fallback = ANSIdefault
		Color.grey = ANSIgrey
		Color.end = ANSIend

		tag.debug = Color.blue + DEBUG + Color.end
		tag.info = Color.green + INFO + Color.end
		tag.warn = Color.yellow + WARN + Color.end
		tag.err = Color.red + ERROR + Color.end
		tag.fatal = Color.red + FATAL + Color.end
	} else {
		Color.black = ""
		Color.red = ""
		Color.green = ""
		Color.yellow = ""
		Color.blue = ""
		Color.magenta = ""
		Color.cyan = ""
		Color.white = ""
		Color.fallback = ""
		Color.grey = ""
		Color.end = ""

		tag.debug = DEBUG
		tag.info = INFO
		tag.warn = WARN
		tag.err = ERROR
		tag.fatal = FATAL
	}
}

// Timestamp returns the current local timestamp as a formatted string
func Timestamp() string {
	return fmt.Sprintf("%s[%s]%s", Color.grey, time.Now().Format(timestampFormat), Color.end)
}

// sprint formats a log string with [source] if provided otherwise returns given string
func sprint(s string) string {
	start := strings.Index(s, "[")
	end := strings.Index(s, "]")
	if start != -1 && end != -1 {
		if source := s[start+1:end] + ":"; source != ":" {
			s = fmt.Sprintf("%s%s%s %s", Color.cyan, source, Color.end, s[end+2:])
		}
	}

	return s
}

// printf formats log text with a message tag and prints to console
func printf(tag, text string) {
	fmt.Printf("%s  %s %s", Timestamp(), tag, sprint(text))
}

// Debugf prints a format specified string with DEBUG message tag if globals.Arguments["--debug"]
func Debugf(format string, a ...any) {
	if globals.Arguments["--debug"] != nil && globals.Arguments["--debug"].(bool) {
		text := fmt.Sprintf(format, a...)
		printf(tag.debug, text)
	}
}

// Printf prints a format specified string with INFO message tag
func Printf(format string, a ...any) {
	text := fmt.Sprintf(format, a...)
	printf(tag.info, text)
}

// Warnf prints a format specified string with WARN message tag
func Warnf(format string, a ...any) {
	text := fmt.Sprintf(format, a...)
	printf(tag.warn, text)
}

// Errorf prints a format specified string with ERROR message tag
func Errorf(format string, a ...any) {
	text := fmt.Sprintf(format, a...)
	printf(tag.err, text)
}

// Fatalf prints a format specified string with FATAL message tag followed by a call to [os.Exit](1)
func Fatalf(format string, a ...any) {
	text := fmt.Sprintf(format, a...)
	printf(tag.fatal, text)
	os.Exit(1)
}
