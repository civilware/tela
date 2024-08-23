package logger

// ANSI colors
const (
	ANSIblack   = "\033[30m"
	ANSIred     = "\033[31m"
	ANSIgreen   = "\033[32m"
	ANSIyellow  = "\033[33m"
	ANSIblue    = "\033[34m"
	ANSImagenta = "\033[35m"
	ANSIcyan    = "\033[36m"
	ANSIwhite   = "\033[37m"
	ANSIdefault = "\033[39m"
	ANSIgrey    = "\033[90m"
	ANSIend     = "\033[0m"
)

type colors struct {
	black    string
	red      string
	green    string
	yellow   string
	blue     string
	magenta  string
	cyan     string
	white    string
	fallback string
	grey     string
	end      string
}

// Colors for logger package, enabled or disabled with EnableColors()
var Color = colors{
	black:    ANSIblack,
	red:      ANSIred,
	green:    ANSIgreen,
	yellow:   ANSIyellow,
	blue:     ANSIblue,
	magenta:  ANSImagenta,
	cyan:     ANSIcyan,
	white:    ANSIwhite,
	fallback: ANSIdefault,
	grey:     ANSIgrey,
	end:      ANSIend,
}

// ANSI color black
func (c *colors) Black() string {
	return c.black
}

// ANSI color red
func (c *colors) Red() string {
	return c.red
}

// ANSI color green
func (c *colors) Green() string {
	return c.green
}

// ANSI color yellow
func (c *colors) Yellow() string {
	return c.yellow
}

// ANSI color blue
func (c *colors) Blue() string {
	return c.blue
}

// ANSI color magenta
func (c *colors) Magenta() string {
	return c.magenta
}

// ANSI color cyan
func (c *colors) Cyan() string {
	return c.cyan
}

// ANSI color white
func (c *colors) White() string {
	return c.white
}

// ANSI color default
func (c *colors) Default() string {
	return c.fallback
}

// ANSI color grey
func (c *colors) Grey() string {
	return c.grey
}

// ANSI color end
func (c *colors) End() string {
	return c.end
}
