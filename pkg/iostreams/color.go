package iostreams

import "github.com/fatih/color"

// Color helpers wrap a string in ANSI color codes when colorEnabled is true.
// When disabled (non-TTY, NO_COLOR set, or --no-color), they return the input
// unchanged so piped consumers see plain text.
//
// We use fatih/color because it handles NO_COLOR detection and Windows ANSI
// emulation. We construct per-call *color.Color instances and explicitly call
// EnableColor() to bypass the package-global NoColor flag — IOStreams owns
// the decision of whether to emit color, not fatih/color's auto-detection.

func colorize(c *color.Color, str string) string {
	c.EnableColor()
	return c.Sprintf("%s", str)
}

func (s *IOStreams) ColorGreen(str string) string {
	if !s.colorEnabled {
		return str
	}
	return colorize(color.New(color.FgGreen), str)
}

func (s *IOStreams) ColorRed(str string) string {
	if !s.colorEnabled {
		return str
	}
	return colorize(color.New(color.FgRed), str)
}

func (s *IOStreams) ColorMagenta(str string) string {
	if !s.colorEnabled {
		return str
	}
	return colorize(color.New(color.FgMagenta), str)
}
