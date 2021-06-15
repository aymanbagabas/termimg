package termimg

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
)

type itermPrinter struct{}

func (p *itermPrinter) PrintTo(w io.Writer, buf *bytes.Buffer, cfg *Config) error {
	out := ""
	b64str := base64.StdEncoding.EncodeToString(buf.Bytes())

	if isTmux() {
		out += "\x1bPtmux;\x1b\x1b]"
	} else {
		out += "\x1b]"
	}

	out += "1337;File=inline=1;preserveAspectRatio="
	if cfg.PreserveAspectRatio {
		out += "1"
	} else {
		out += "0"
	}

	out += fmt.Sprintf(";size=%d", buf.Len())
	if cfg.Width != nil {
		out += fmt.Sprintf(";width=%d", *cfg.Width)
	}

	if cfg.Height != nil {
		out += fmt.Sprintf(";height=%d", *cfg.Height)
	}

	out += fmt.Sprintf(":%s", b64str)
	if isTmux() {
		out += "\x07\x1b\\"
	} else {
		out += "\x07"
	}

	_, err := fmt.Fprintln(w, out)

	return err
}
