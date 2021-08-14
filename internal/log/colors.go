package log

import "fmt"

const (
	ansi_reset  = "\033[0m"
	ansi_green  = "\033[32m"
	ansi_yellow = "\033[33m"
	ansi_blue   = "\033[34m"
	ansi_purple = "\033[35m"
)

// Color fmt format for color.
type Color string

// Green color.
func Green() Color {
	return Color(fmt.Sprintf("%s%%s%s", ansi_green, ansi_reset))
}

// Yellow color.
func Yellow() Color {
	return Color(fmt.Sprintf("%s%%s%s", ansi_yellow, ansi_reset))
}

// Blue color.
func Blue() Color {
	return Color(fmt.Sprintf("%s%%s%s", ansi_blue, ansi_reset))
}

// Purple color.
func Purple() Color {
	return Color(fmt.Sprintf("%s%%s%s", ansi_purple, ansi_reset))
}

// Clean color.
func Clean() Color {
	return "%s"
}
