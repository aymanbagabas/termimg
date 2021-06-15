package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aymanbagabas/termimg"
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
			cfg := termimg.NewConfig()
			cfg.AbsoluteOffset = Absolute
			cfg.PreserveAspectRatio = PreserveAspectRatio
			cfg.X = uint(X)
			cfg.Y = Y

			if Backend != "" {
				backend := strings.ToLower(Backend)
				cfg.UseIterm = false
				cfg.UseKitty = false
				cfg.UseSixel = false
				cfg.UseBlocks = false

				switch backend {
				case "iterm":
					cfg.UseIterm = true
				case "kitty":
					cfg.UseKitty = true
				case "sixel":
					cfg.UseSixel = true
				case "blocks":
					cfg.UseBlocks = true
				default:
					return errors.New("unknown backend")
				}
			}

			if Width > 0 {
				cfg.Width = &Width
			}
			if Height > 0 {
				cfg.Height = &Height
			}

			return termimg.PrintFromFile(args[0], cfg)
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
