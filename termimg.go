package termimg

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"syscall"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"golang.org/x/sys/unix"
)

type PrinterType string

const (
	Iterm  PrinterType = "iterm"
	Kitty  PrinterType = "kitty"
	Sixel  PrinterType = "sixel"
	Blocks PrinterType = "blocks"
)

var (
	ErrPrinterNotSupported = errors.New("printer not supported")
	ErrUnknownPrinter      = errors.New("unknown printer")
	ErrUnknownWinSize      = errors.New("unknown window size")
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
		UseIterm:            false,
		UseKitty:            false,
		UseSixel:            false,
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
		fmt.Fprintf(w, "\x1b[%d;%dH", cfg.Y, cfg.X)
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
	} else if cfg.UseBlocks {
		log.Println("printer: blocks")
		printer = &blocksPrinter{}
	} else {
		return ErrPrinterNotSupported
	}

	return printer.PrintTo(w, buf, cfg)
}

func getWinSize(w io.Writer) (*unix.Winsize, error) {
	var outputAsFile *os.File
	if f, ok := w.(*os.File); ok {
		outputAsFile = f
	}

	if outputAsFile != nil {
		fd := outputAsFile.Fd()
		ws, err := unix.IoctlGetWinsize(int(fd), syscall.TIOCGWINSZ)
		if err != nil {
			return nil, err
		}

		return ws, nil
	} else {
		return nil, ErrUnknownWinSize
	}
}

func isTmux() bool {
	return strings.Contains(strings.ToLower(os.Getenv("TERM_PROGRAM")), "tmux")
}
