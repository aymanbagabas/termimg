package termimg

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/containerd/console"
	te "github.com/muesli/termenv"
	"golang.org/x/term"
)

type PrinterType string

const (
	Iterm  PrinterType = "iterm"
	Kitty  PrinterType = "kitty"
	Sixel  PrinterType = "sixel"
	Blocks PrinterType = "blocks"
)

type Config struct {
	X                   uint
	Y                   int
	Width               *int
	Height              *int
	AbsoluteOffset      bool
	PreserveAspectRatio bool
	UseIterm            bool
	UseKitty            bool
	UseSixel            bool
	UseBlocks           bool
}

type Printer interface {
	PrintTo(w io.Writer, buf *bytes.Buffer, cfg *Config) error
}

func (p PrinterType) String() string {
	return string(p)
}

func NewConfig() *Config {
	return &Config{
		X:                   0,
		Y:                   0,
		Width:               nil,
		Height:              nil,
		AbsoluteOffset:      false,
		PreserveAspectRatio: true,
		UseIterm:            true,
		UseKitty:            true,
		UseSixel:            true,
		UseBlocks:           true,
	}
}

func PrintFromFile(path string, cfg *Config) error {
	return PrintFromFileTo(os.Stdout, path, cfg)
}

func PrintFromFileTo(w io.Writer, path string, cfg *Config) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(f)
	if err != nil {
		return err
	}

	return PrintTo(w, buf, cfg)
}
func Print(buf *bytes.Buffer, cfg *Config) error {
	return PrintTo(os.Stdout, buf, cfg)
}

func PrintTo(w io.Writer, buf *bytes.Buffer, cfg *Config) error {
	var printer Printer

	if cfg == nil {
		cfg = &Config{}
	}

	if cfg.AbsoluteOffset {
		fmt.Fprintf(w, te.CSI+te.CursorPositionSeq, cfg.Y, cfg.X)
	}

	log.Printf("%+v\n", cfg)

	if cfg.UseIterm && supportsIterm() {
		log.Println("printer: iterm")
		printer = &itermPrinter{}
	} else if cfg.UseKitty && supportsKitty() && !isTmux() {
		log.Println("printer: kitty")
		printer = &kittyPrinter{}
	} else if cfg.UseSixel && supportsSixel() && !isTmux() {
		log.Println("printer: sixel")
		printer = &sixelPrinter{}
	} else {
		log.Println("printer: blocks")
		printer = &blocksPrinter{}
	}

	return printer.PrintTo(w, buf, cfg)
}

func getTermBounds(w io.Writer) (int, int) {
	// Assume default terminal size
	cols := 80
	lines := 24

	var outputAsFile *os.File
	if f, ok := w.(*os.File); ok {
		outputAsFile = f
	}

	if outputAsFile != nil {
		fd := int(outputAsFile.Fd())
		if term.IsTerminal(fd) {
			w, h, err := term.GetSize(fd)
			if err == nil {
				cols = w
				lines = h
			}
		}
	}

	return cols, lines
}

func supportsIterm() bool {
	return os.Getenv("ITERM_SESSION_ID") != ""
}

func supportsKitty() bool {
	return false
}

func supportsSixelTermAttrs() bool {
	resp := make([]byte, 0)
	out := make([]byte, 1)
	c := console.Current()

	_, err := c.Write([]byte(te.CSI + "c"))
	if err != nil {
		log.Fatal(err.Error())
		return false
	}

	for _, err := c.Read(out); err != io.EOF; {
		resp = append(resp, out[0])
		if out[0] == 'c' {
			break
		}
	}

	str := string(resp)

	return strings.Contains(str, ";4;") || strings.Contains(str, ";4c")
}

func supportsSixel() bool {
	env := os.Getenv("TERM")
	prog := os.Getenv("TERM_PROGRAM")

	if prog == "MacTerm" {
		return true
	}

	switch env {
	case "mlterm", "yaft-256color", "foot":
		return true
	case "st-256color", "xterm", "xterm-256color":
		return supportsSixelTermAttrs()
	}

	return false
}

func isTmux() bool {
	return strings.Contains(strings.ToLower(os.Getenv("TERM_PROGRAM")), "tmux")
}
