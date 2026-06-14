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
	"github.com/djmelvee/golazo-tui/internal/wc"
)

func main() {
	watch := flag.Bool("watch", false, "keep running and poll on an interval")
	interval := flag.Int("interval", 60, "polling interval in seconds (used with --watch)")
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

	fetch := func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		live, err := client.FetchMatches(ctx, "live")
		if err != nil {
			return fmt.Errorf("fetch live: %w", err)
		}
		if err := db.Set("matches:live", live); err != nil {
			return fmt.Errorf("set live: %w", err)
		}

		upcoming, err := client.FetchMatches(ctx, "upcoming")
		if err != nil {
			return fmt.Errorf("fetch upcoming: %w", err)
		}
		if err := db.Set("matches:upcoming", upcoming); err != nil {
			return fmt.Errorf("set upcoming: %w", err)
		}

		finished, err := client.FetchMatches(ctx, "finished")
		if err != nil {
			return fmt.Errorf("fetch finished: %w", err)
		}
		if err := db.Set("matches:finished", finished); err != nil {
			return fmt.Errorf("set finished: %w", err)
		}

		standings, err := client.FetchStandings(ctx)
		if err != nil {
			return fmt.Errorf("fetch standings: %w", err)
		}
		if err := db.Set("standings", standings); err != nil {
			return fmt.Errorf("set standings: %w", err)
		}

		fmt.Printf("[%s] Fetched %d live, %d upcoming, %d finished matches\n",
			time.Now().Format("15:04:05"), len(live), len(upcoming), len(finished))
		return nil
	}

	if err := fetch(); err != nil {
		log.Printf("fetch error: %v", err)
	}

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
			if err := fetch(); err != nil {
				log.Printf("fetch error: %v", err)
			}
		case <-sig:
			fmt.Println("\nStopped.")
			return
		}
	}
}
