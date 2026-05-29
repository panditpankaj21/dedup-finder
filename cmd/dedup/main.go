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

    paths := make(chan string, 100)
    hashes := make(map[string][]string)
    var mu sync.Mutex

    // Workers run in a goroutine so main can also walk
    var workerWg sync.WaitGroup
    workerWg.Add(1)
    go func() {
        defer workerWg.Done()
        worker.Pool(ctx, numWorkers, paths, hashes, &mu)
    }()

    files, err := walker.Walk(ctx, root, paths)
    close(paths)
    if err != nil {
        log.Printf("WARN: walk: %v", err)
    }

    workerWg.Wait()

    elapsed := time.Since(start)
    if ctx.Err() != nil {
        fmt.Printf("\nCancelled — partial results below\n")
    }
    fmt.Printf("Took %v\n\nScanned %d files\n", elapsed, files)

    groups := report.Print(hashes)
    if groups == 0 {
        fmt.Println("\nNo duplicates found.")
    } else {
        fmt.Printf("\nFound %d duplicate group(s).\n", groups)
    }
}