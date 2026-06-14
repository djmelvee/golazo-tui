// golazo-fetcher polls the WC2026 REST API and writes results to the
// local SQLite cache. Run once or with --watch for continuous updates.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/djmelvee/golazo-tui/internal/data"
	"github.com/djmelvee/golazo-tui/internal/fetcher"
	"github.com/djmelvee/golazo-tui/internal/wc"
)

func main() {
	watch := flag.Bool("watch", false, "keep running and poll on an interval")
	interval := flag.Int("interval", 5, "polling interval in seconds (used with --watch)")
	flag.Parse()

	dbPath := os.Getenv("GOLAZO_DB")
	if dbPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
		dbPath = filepath.Join(home, ".cache", "golazo-tui", "cache.db")
	}

	apiBase := os.Getenv("GOLAZO_API")
	if apiBase == "" {
		apiBase = "http://worldcup26.ir:3050"
	}

	apiToken := os.Getenv("GOLAZO_API_TOKEN")
	if apiToken == "" {
		fmt.Fprintln(os.Stderr, "Error: GOLAZO_API_TOKEN is not set.")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "To get a token (free, one-time registration):")
		fmt.Fprintln(os.Stderr, `  curl -X POST http://worldcup26.ir:3050/auth/register \`)
		fmt.Fprintln(os.Stderr, `       -H "Content-Type: application/json" \`)
		fmt.Fprintln(os.Stderr, `       -d '{"username":"<you>","password":"<pass>","email":"<email>"}'`)
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Then:")
		fmt.Fprintln(os.Stderr, "  export GOLAZO_API_TOKEN=<token>")
		fmt.Fprintln(os.Stderr, "  golazo-fetcher --watch")
		os.Exit(1)
	}

	db, err := data.Open(dbPath)
	if err != nil {
		log.Fatalf("open DB: %v", err)
	}
	defer db.Close()

	client := wc.New(apiBase, apiToken)

	do := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := fetcher.Fetch(ctx, client, db); err != nil {
			log.Printf("fetch error: %v", err)
		} else {
			fmt.Printf("[%s] Fetched OK\n", time.Now().Format("15:04:05"))
		}
	}

	do()

	if !*watch {
		return
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(time.Duration(*interval) * time.Second)
	defer ticker.Stop()

	fmt.Printf("Watching — polling every %ds. Ctrl+C to stop.\n", *interval)
	for {
		select {
		case <-ticker.C:
			do()
		case <-sig:
			fmt.Println("\nStopped.")
			return
		}
	}
}
