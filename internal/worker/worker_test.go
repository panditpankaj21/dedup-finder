package worker

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestPool(t *testing.T) {
	t.Run("processes all files in channel", func(t *testing.T) {
		// ARRANGE: create 3 files with KNOWN duplicate pattern
		// Two files share content "same", one has unique "unique".
		// After processing, the map should have:
		//   - one hash mapped to 2 paths (the duplicates)
		//   - one hash mapped to 1 path (the unique)
		dir := t.TempDir()
		dup1 := filepath.Join(dir, "dup1.txt")
		dup2 := filepath.Join(dir, "dup2.txt")
		unique := filepath.Join(dir, "unique.txt")
		os.WriteFile(dup1, []byte("same"), 0644)
		os.WriteFile(dup2, []byte("same"), 0644)
		os.WriteFile(unique, []byte("unique"), 0644)

		// ACT: feed paths through channel, run pool
		paths := make(chan string, 10)
		paths <- dup1
		paths <- dup2
		paths <- unique
		close(paths) // signals workers no more work coming

		hashes := make(map[string][]string)
		var mu sync.Mutex

		// Pool blocks until all workers exit
		Pool(context.Background(), 2, paths, hashes, &mu)

		// ASSERT
		if len(hashes) != 2 {
			t.Errorf("expected 2 unique hashes, got %d", len(hashes))
		}

		// Find the duplicate group and verify it has 2 paths
		foundDuplicate := false
		for _, paths := range hashes {
			if len(paths) == 2 {
				foundDuplicate = true
			}
		}
		if !foundDuplicate {
			t.Error("expected one hash group with 2 paths (duplicates), found none")
		}
	})

	t.Run("exits when channel closes", func(t *testing.T) {
		// ARRANGE: empty channel that closes immediately
		// Workers should exit cleanly without processing anything.
		paths := make(chan string)
		close(paths)

		hashes := make(map[string][]string)
		var mu sync.Mutex

		// ACT: Pool returns when all workers exit. If broken, test hangs.
		// We add a timeout in test runner via `-timeout` flag (default 10min).
		Pool(context.Background(), 4, paths, hashes, &mu)

		// ASSERT: no work done, map stays empty
		if len(hashes) != 0 {
			t.Errorf("expected empty map, got %d entries", len(hashes))
		}
	})

	t.Run("exits when context cancelled", func(t *testing.T) {
		// ARRANGE: cancelled context, channel with items
		// Workers should see ctx.Done() and exit without processing.
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		paths := make(chan string, 10)
		paths <- "some-path" // intentionally unprocessable, but workers won't get to it
		close(paths)

		hashes := make(map[string][]string)
		var mu sync.Mutex

		// ACT: should return quickly because ctx is already cancelled
		// We wrap in a goroutine + timeout to fail-fast if Pool hangs
		done := make(chan struct{})
		go func() {
			Pool(ctx, 4, paths, hashes, &mu)
			close(done)
		}()

		select {
		case <-done:
			// Pool returned in time
		case <-time.After(2 * time.Second):
			t.Fatal("Pool did not return within 2s after context cancellation")
		}
	})
}