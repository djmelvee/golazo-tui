// golazo-tui — FIFA World Cup 2026 terminal dashboard
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/djmelvee/golazo-tui/internal/app"
	"github.com/djmelvee/golazo-tui/internal/auth"
	"github.com/djmelvee/golazo-tui/internal/data"
	"github.com/djmelvee/golazo-tui/internal/wc"
)

const version = "v0.6.0"

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

	db, err := data.Open(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot open cache DB at %s: %v\n", dbPath, err)
		os.Exit(1)
	}
	defer db.Close()

	apiBase := os.Getenv("GOLAZO_API")
	if apiBase == "" {
		apiBase = "http://worldcup26.ir:3050"
	}

	// Resolve API token: env var → stored in DB → auto-register on first launch.
	token := os.Getenv("GOLAZO_API_TOKEN")
	if token == "" {
		token = db.GetToken()
	}
	var startNote string
	if token == "" {
		ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
		t, regErr := auth.Register(ctx, apiBase)
		cancel()
		if regErr == nil {
			token = t
			_ = db.SetToken(token)
		} else {
			// Surface the raw error in the TUI header and write to a debug log.
			startNote = "auth: " + regErr.Error()
			logPath := filepath.Join(filepath.Dir(dbPath), "debug.log")
			_ = os.WriteFile(logPath, []byte("registration error: "+regErr.Error()+"\n"), 0644)
		}
	}

	// Always create a client — if token is empty, the API may allow unauthed reads;
	// if not, fetch errors appear in the header instead of silent offline mode.
	client := wc.New(apiBase, token)

	model := app.New(db, client)
	if startNote != "" {
		model = model.WithNote(startNote)
	}
	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
