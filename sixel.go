package termimg

import (
	"bufio"
	"bytes"
	"fmt"
	"image"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"image/color"
	"image/color/palette"
	"image/draw"
	"image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/mattn/go-sixel"
)

type sixelPrinter struct{}

func (p *sixelPrinter) PrintTo(w io.Writer, buf *bytes.Buffer, cfg *Config) error {
	enc := sixel.NewEncoder(w)

	contentType := http.DetectContentType(buf.Bytes())
	if contentType == "image/gif" {
		return p.printGif(w, buf, cfg)
	}

	img, _, err := image.Decode(buf)
	if err != nil {
		return err
	}

	return enc.Encode(img)
}

func (p *sixelPrinter) printGif(w io.Writer, buf *bytes.Buffer, cfg *Config) error {
	g, err := gif.DecodeAll(buf)
	if err != nil {
		return err
	}

	enc := sixel.NewEncoder(w)
	enc.Width = g.Config.Width
	enc.Height = g.Config.Height
	ws, err := getWinSize(w)
	log.Printf("winsize: %+v\n", ws)
	if err != nil {
		return err
	}

	if ws.Xpixel > 0 && ws.Ypixel > 0 && ws.Col > 0 && ws.Row > 0 {
		height := float64(ws.Ypixel) / float64(ws.Row)
		lines := int(math.Ceil(float64(enc.Height) / height))
		fmt.Fprint(w, strings.Repeat("\n", lines))
		fmt.Fprintf(w, "\x1b[%dA", lines)
		fmt.Fprint(w, "\x1b[s")
	}

	paletteFactor := append(palette.WebSafe, color.Transparent)
	bounds := g.Image[0].Bounds()
	for {
		t := time.Now()
		for i, frame := range g.Image {
			fmt.Fprint(w, "\x1b[u")
			paletteImage := image.NewPaletted(bounds, paletteFactor)
			draw.Draw(paletteImage, bounds, &image.Uniform{frame.Palette[0]}, image.Pt(0, 0), draw.Src)
			draw.Draw(paletteImage, bounds, frame, image.Pt(0, 0), draw.Src)
			enc.Encode(paletteImage)
			span := time.Second * time.Duration(g.Delay[i]) / 100
			if time.Since(t) < span {
				time.Sleep(span)
			}
			t = time.Now()
		}
		if g.LoopCount != 0 {
			g.LoopCount--
			if g.LoopCount == 0 {
				break
			}
		}
	}

	return nil
}

func supportsSixelTermAttrs() bool {
	var resp string
	_, w, _ := os.Pipe()
	f := os.Stdout
	os.Stdout = w
	out := make([]byte, 1)
	wr := bufio.NewWriter(f)
	wr.WriteString("\x1b[c")
	wr.Flush()

	for {
		_, err := os.Stdin.Read(out)
		if err != nil {
			log.Fatal(err.Error())
			break
		}

		resp += string(out[0])
		if out[0] == 'c' || err == io.EOF {
			break
		}
	}

	return strings.Contains(resp, ";4;") || strings.Contains(resp, ";4c")
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
