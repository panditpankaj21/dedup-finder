package walker

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestWalk(t *testing.T) {
	t.Run("walks all files", func(t *testing.T) {
		// ARRANGE: create temp dir with 3 files
		dir := t.TempDir()
		for _, name := range []string{"a.txt", "b.txt", "c.txt"} {
			if err := os.WriteFile(filepath.Join(dir, name), []byte("data"), 0644); err != nil {
				t.Fatalf("failed to write %s: %v", name, err)
			}
		}

		// ACT: run Walk and drain the channel
		// Note: buffered channel sized larger than file count so Walk never blocks.
		// In production code we drain in a goroutine — see Pattern A in real worker pools.
		paths := make(chan string, 100)
		count, err := Walk(context.Background(), dir, paths)
		close(paths)

		// Drain whatever Walk sent
		var received []string
		for p := range paths {
			received = append(received, p)
		}

		// ASSERT
		if err != nil {
			t.Fatalf("walk returned error: %v", err)
		}
		if count != 3 {
			t.Errorf("expected count=3, got %d", count)
		}
		if len(received) != 3 {
			t.Errorf("expected 3 paths in channel, got %d", len(received))
		}
	})

	t.Run("respects cancellation", func(t *testing.T) {
		// ARRANGE: create files
		dir := t.TempDir()
		for _, name := range []string{"a.txt", "b.txt", "c.txt"} {
			os.WriteFile(filepath.Join(dir, name), []byte("data"), 0644)
		}

		// Cancel the context BEFORE walking.
		// Walk's first send-attempt sees ctx.Done() ready and returns filepath.SkipAll.
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		// ACT
		paths := make(chan string, 100)
		count, _ := Walk(ctx, dir, paths)
		close(paths)

		// ASSERT: nothing was queued because we cancelled first
		if count != 0 {
			t.Errorf("expected count=0 (cancelled), got %d", count)
		}
	})

	t.Run("walks nested directories", func(t *testing.T) {
		// ARRANGE: create files in subdirectories — proves recursion works
		dir := t.TempDir()
		subDir := filepath.Join(dir, "nested")
		if err := os.MkdirAll(subDir, 0755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		os.WriteFile(filepath.Join(dir, "top.txt"), []byte("a"), 0644)
		os.WriteFile(filepath.Join(subDir, "deep.txt"), []byte("b"), 0644)

		// ACT
		paths := make(chan string, 100)
		count, err := Walk(context.Background(), dir, paths)
		close(paths)

		// ASSERT
		if err != nil {
			t.Fatalf("walk error: %v", err)
		}
		if count != 2 {
			t.Errorf("expected count=2, got %d", count)
		}
	})
}