package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"dedup-finder/internal/report"
	"dedup-finder/internal/sizegroup"
	"dedup-finder/internal/walker"
	"dedup-finder/internal/worker"
)

func main() {
	workers := flag.Int("workers", 8, "number of concurrent workers")
	minSize := flag.Int64("min-size", 0, "skip files smaller than N bytes")
	ignore := flag.String("ignore", "", "comma-separated patterns to skip (e.g., '.git,*.log')")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] <directory>\n\nFlags:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	log.Println(*workers)

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}	
	root := flag.Arg(0)

	var ignorePatterns []string
	if *ignore != "" {
		ignorePatterns = strings.Split(*ignore, ",")
		for i := range ignorePatterns {
			ignorePatterns[i] = strings.TrimSpace(ignorePatterns[i])
		}
	}

	log.Printf("Scanning %s (workers=%d, min-size=%d, ignore=%v)...", root, *workers, *minSize, ignorePatterns)
	start := time.Now()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Printf("Cancellation requested, stopping...")
		cancel()
	}()

	walkerOut := make(chan string, 100)
	go func() {
		defer close(walkerOut)
		walker.Walk(ctx, root, ignorePatterns, walkerOut)
	}()

	candidates, totalSeen := sizegroup.Filter(ctx, walkerOut, *minSize)

	log.Printf("Stat phase: %d files seen, %d candidates need hashing", totalSeen, len(candidates))

	workerIn := make(chan string, 100)
	go func() {
		defer close(workerIn)
		for _, p := range candidates {
			select {
			case <-ctx.Done():
				return
			case workerIn <- p:
			}
		}
	}()

	hashes := make(map[string][]string)
	var mu sync.Mutex
	worker.Pool(ctx, *workers, workerIn, hashes, &mu)

	elapsed := time.Since(start)
	if ctx.Err() != nil {
		fmt.Printf("\nCancelled — partial results below\n")
	}
	fmt.Printf("Took %v\n\nScanned %d files\n", elapsed, totalSeen)

	groups := report.Print(hashes)
	if groups == 0 {
		fmt.Println("\nNo duplicates found.")
	} else {
		fmt.Printf("\nFound %d duplicate group(s).\n", groups)
	}
}
