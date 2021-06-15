package termimg

import (
	"bytes"
	"io"
)

type blocksPrinter struct{}

func (p *blocksPrinter) PrintTo(w io.Writer, buf *bytes.Buffer, cfg *Config) error {
	return nil
}
