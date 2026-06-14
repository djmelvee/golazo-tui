// golazo-tui — FIFA World Cup 2026 terminal dashboard
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"

	"github.com/djmelvee/golazo-tui/internal/app"
	"github.com/djmelvee/golazo-tui/internal/data"
)

const version = "v0.3.0"

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Println("golazo-tui " + version)
		return
	}

	dbPath := os.Getenv("GOLAZO_DB")
	if dbPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
		dbPath = filepath.Join(home, ".cache", "golazo-tui", "cache.db")
	}

	db, err := data.OpenRO(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot open cache DB at %s: %v\n", dbPath, err)
		fmt.Fprintf(os.Stderr, "Run: go run ./cmd/golazo-seed\n")
		os.Exit(1)
	}
	defer db.Close()

	p := tea.NewProgram(app.New(db))
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
