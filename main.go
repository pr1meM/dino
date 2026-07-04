package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/gdamore/tcell/v2"

	"dino/internal/buffer"
	"dino/internal/config"
	"dino/internal/editor"
)

func main() {
	tabSize := flag.Int("tabsize", 0, "indent size in spaces (overrides the config file)")
	useTabs := flag.Bool("tabs", false, "use real tabs instead of spaces for indentation")
	flag.Parse()

	cfg, err := config.Load(config.Path())
	if err != nil {
		fmt.Fprintln(os.Stderr, "dino: warning: could not load config:", err)
	}
	if *tabSize > 0 {
		cfg.TabSize = *tabSize
	}
	if *useTabs {
		cfg.UseSpaces = false
	}

	screen, err := tcell.NewScreen()
	if err != nil {
		fmt.Fprintln(os.Stderr, "dino: could not initialize terminal:", err)
		os.Exit(1)
	}
	if err := screen.Init(); err != nil {
		fmt.Fprintln(os.Stderr, "dino: could not initialize terminal:", err)
		os.Exit(1)
	}
	screen.EnableMouse()
	screen.SetTitle("Dino")
	defer screen.Fini()

	ed := editor.New(screen, editor.Config{TabSize: cfg.TabSize, UseSpaces: cfg.UseSpaces})

	for _, path := range flag.Args() {
		if err := ed.OpenFile(path); err != nil {
			if os.IsNotExist(err) {
				b := buffer.New()
				b.FilePath = path
				ed.OpenBuffer(b)
			} else {
				screen.Fini()
				fmt.Fprintln(os.Stderr, "dino:", err)
				os.Exit(1)
			}
		}
	}

	if err := ed.Run(); err != nil {
		screen.Fini()
		fmt.Fprintln(os.Stderr, "dino:", err)
		os.Exit(1)
	}
}
