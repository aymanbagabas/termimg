package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	ti "github.com/aymanbagabas/termimg"
	"github.com/spf13/cobra"
)

var (
	Backend             string
	Absolute            bool
	PreserveAspectRatio bool
	X                   int
	Y                   int
	Width               int
	Height              int
)

var (
	rootCmd = &cobra.Command{
		Use:   "teimg [flags] file",
		Short: "Display images in the terminal",
		Args:  cobra.ExactArgs(1),

		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := ti.NewConfig()
			cfg.AbsoluteOffset = Absolute
			cfg.PreserveAspectRatio = PreserveAspectRatio
			cfg.X = uint(X)
			cfg.Y = Y

			backend := strings.ToLower(Backend)
			cfg.UseIterm = backend == "" || backend == ti.Iterm.String()
			cfg.UseKitty = backend == "" || backend == ti.Kitty.String()
			cfg.UseSixel = backend == "" || backend == ti.Sixel.String()
			cfg.UseBlocks = backend == "" || backend == ti.Blocks.String()

			if backend != "" && (backend != ti.Iterm.String() &&
				backend != ti.Kitty.String() &&
				backend != ti.Sixel.String() &&
				backend != ti.Blocks.String()) {
				return ti.ErrUnknownPrinter
			}

			if Width > 0 {
				cfg.Width = &Width
			}
			if Height > 0 {
				cfg.Height = &Height
			}

			return ti.PrintFromFile(args[0], cfg)
		},
	}
)

func main() {
	f, err := os.OpenFile("debug.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)

	rootCmd.Flags().Bool("help", false, "Print usage")
	rootCmd.Flags().StringVarP(&Backend, "backend", "b", "", "Backend to render image. Available backends: [iterm, kitty, sixel, blocks]")
	rootCmd.Flags().BoolVarP(&Absolute, "absolute", "a", false, "Absolute positioning")
	rootCmd.Flags().BoolVarP(&PreserveAspectRatio, "preserveRatio", "p", true, "Preserve image aspect ratio")
	rootCmd.Flags().IntVarP(&X, "x", "x", 0, "X position")
	rootCmd.Flags().IntVarP(&Y, "y", "y", 0, "Y position")
	rootCmd.Flags().IntVarP(&Width, "width", "w", 0, "Image width")
	rootCmd.Flags().IntVarP(&Height, "height", "h", 0, "Image height")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
