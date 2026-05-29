package worker

import (
    "context"
    "log"
    "sync"

    "dedup-finder/internal/hasher"
)


func Pool(ctx context.Context, n int, paths <-chan string, hashes map[string][]string, mu *sync.Mutex) {
    var wg sync.WaitGroup
    for i := 0; i < n; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for {
                select {
                case <-ctx.Done():
                    return
                case path, ok := <-paths:
                    if !ok {
                        return
                    }
                    h, err := hasher.File(path)
                    if err != nil {
                        log.Printf("WARN: hash %s: %v", path, err)
                        continue
                    }
                    mu.Lock()
                    hashes[h] = append(hashes[h], path)
                    mu.Unlock()
                }
            }
        }()
    }
    wg.Wait()
}