package termimg

import (
	"bytes"
	"io"
)

type kittyPrinter struct{}

func (p *kittyPrinter) PrintTo(w io.Writer, buf *bytes.Buffer, cfg *Config) error {
	return nil
}
