package termimg

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

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
	} else {

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

	_, height := getTermBounds(w)
	lines := int(math.Ceil(float64(enc.Height) / float64(height)))
	fmt.Fprint(w, strings.Repeat("\n", lines))
	fmt.Fprintf(w, "\x1b[%dA", lines)
	fmt.Fprintf(w, "\x1b[s")

	var back draw.Image
	if g.BackgroundIndex == 0 {
		back = image.NewPaletted(g.Image[0].Bounds(), palette.WebSafe)
	}
	// log.Printf("back index: %d, %s\n", g.BackgroundIndex, back)

	for {
		t := time.Now()
		for j := 0; j < len(g.Image); j++ {
			// log.Printf("back index: %d, %s\n", g.BackgroundIndex, back)
			fmt.Fprint(w, "\x1b[u")
			if back != nil {
				// log.Println("here")
				draw.Draw(back, back.Bounds(), &image.Uniform{g.Image[j].Palette[g.BackgroundIndex]}, image.Pt(0, 0), draw.Src)
				draw.Draw(back, back.Bounds(), g.Image[j], image.Pt(0, 0), draw.Src)
				err = enc.Encode(back)
			} else {
				err = enc.Encode(g.Image[j])
			}
			if err != nil {
				return err
			}
			span := time.Second * time.Duration(g.Delay[j]) / 100
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
