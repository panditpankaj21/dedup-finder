package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"dedup-finder/internal/report"
	"dedup-finder/internal/sizegroup"
	"dedup-finder/internal/walker"
	"dedup-finder/internal/worker"
)

const numWorkers = 8

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: dedup <directory>")
	}
	root := os.Args[1]
	log.Printf("Scanning %s...", root)
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
		walker.Walk(ctx, root, walkerOut)
	}()

	candidates, totalSeen := sizegroup.Filter(ctx, walkerOut)

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
	worker.Pool(ctx, numWorkers, workerIn, hashes, &mu)

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
