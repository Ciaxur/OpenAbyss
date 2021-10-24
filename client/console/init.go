package console

import (
	"github.com/fatih/color"
)

func Init() {
	Error = color.New(color.FgRed).Add(color.Bold)
	Heading = color.New(color.FgYellow).Add(color.Bold)
	Warning = color.New(color.FgYellow).Add(color.Bold)
	Log = color.New()
	Info = color.New(color.FgHiMagenta)
}
