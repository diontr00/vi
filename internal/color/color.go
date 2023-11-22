package color

import (
	"bytes"
	"fmt"
	"github.com/mattn/go-isatty"
	"os"
)

const (
	GreenFg  = "32"
	BlueFg   = "34"
	RedFg    = "31"
	BoldText = "1"
)

var (
	green = makecolor(GreenFg)
	blue  = makecolor(BlueFg)
	red   = makecolor(RedFg)
	bold  = makecolor(BoldText)

	noColor = !isatty.IsTerminal(os.Stdout.Fd()) && !isatty.IsCygwinTerminal(os.Stdout.Fd())
)

type colorize func(interface{}) string

func makecolor(n string) colorize {
	return func(msg interface{}) string {
		b := new(bytes.Buffer)
		b.WriteString("\x1b[")
		b.WriteString(n)
		b.WriteString("m")
		return fmt.Sprintf("%s%v\x1b[0m", b.String(), msg)
	}
}

func Green(format string, args ...any) string {
	m := makeMsg(format, args...)
	if noColor {
		return m
	}
	return green(m)
}

func Blue(format string, args ...any) string {
	m := makeMsg(format, args...)
	if noColor {
		return m
	}
	return blue(m)
}

func Red(format string, args ...any) string {
	m := makeMsg(format, args...)

	if noColor {
		return m
	}
	return red(m)
}

func Bold(format string, args ...any) string {
	m := makeMsg(format, args...)
	return bold(m)
}

func makeMsg(format string, args ...any) (m string) {
	if len(args) == 0 {
		m = format
	} else {
		m = fmt.Sprintf(format, args...)
	}
	return
}
