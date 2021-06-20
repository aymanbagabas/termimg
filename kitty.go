package termimg

import (
	"bytes"
	"io"
	"os"
	"strings"
)

type kittyPrinter struct{}

func (p *kittyPrinter) PrintTo(w io.Writer, buf *bytes.Buffer, cfg *Config) error {
	return nil
}

func supportsKitty() bool {
	return strings.Contains(strings.ToLower(os.Getenv("TERM")), "kitty")
}
