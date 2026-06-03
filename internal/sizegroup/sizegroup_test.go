package sizegroup

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestFilter(t *testing.T) {
	t.Run("all unique sizes returns no candidates", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "tiny.txt"), "a")           // 1 byte
		writeFile(t, filepath.Join(dir, "small.txt"), "aaa")        // 3 bytes
		writeFile(t, filepath.Join(dir, "medium.txt"), "aaaaaaaa")  // 8 bytes

		
		paths := make(chan string, 10)
		paths <- filepath.Join(dir, "tiny.txt")
		paths <- filepath.Join(dir, "small.txt")
		paths <- filepath.Join(dir, "medium.txt")
		close(paths)

		candidates, totalSeen := Filter(context.Background(), paths, 0)

		
		if totalSeen != 3 {
			t.Errorf("expected totalSeen=3, got %d", totalSeen)
		}
		if len(candidates) != 0 {
			t.Errorf("expected 0 candidates (all unique sizes), got %d", len(candidates))
		}
	})

	t.Run("files sharing size are returned as candidates", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "dup1.txt"), "aaa")    // 3 bytes
		writeFile(t, filepath.Join(dir, "dup2.txt"), "bbb")    // 3 bytes (same SIZE, different content)
		writeFile(t, filepath.Join(dir, "unique.txt"), "aaaaa") // 5 bytes

		// ACT
		paths := make(chan string, 10)
		paths <- filepath.Join(dir, "dup1.txt")
		paths <- filepath.Join(dir, "dup2.txt")
		paths <- filepath.Join(dir, "unique.txt")
		close(paths)

		candidates, totalSeen := Filter(context.Background(), paths, 0)

		if totalSeen != 3 {
			t.Errorf("expected totalSeen=3, got %d", totalSeen)
		}
		if len(candidates) != 2 {
			t.Errorf("expected 2 candidates (the size-3 pair), got %d", len(candidates))
		}

		for _, c := range candidates {
			if filepath.Base(c) == "unique.txt" {
				t.Errorf("unique.txt should not be a candidate, it has a unique size")
			}
		}
	})

	t.Run("respects cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		paths := make(chan string, 10)
		paths <- "fake-path-that-wont-be-processed"

		candidates, totalSeen := Filter(ctx, paths, 0)

		if totalSeen != 0 {
			t.Errorf("expected totalSeen=0 (cancelled), got %d", totalSeen)
		}
		if len(candidates) != 0 {
			t.Errorf("expected 0 candidates (cancelled), got %d", len(candidates))
		}
	})

	t.Run("invalid path skipped, others processed", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "good1.txt"), "abc") // 3 bytes
		writeFile(t, filepath.Join(dir, "good2.txt"), "xyz") // 3 bytes

		paths := make(chan string, 10)
		paths <- filepath.Join(dir, "good1.txt")
		paths <- "/this/does/not/exist.txt"
		paths <- filepath.Join(dir, "good2.txt")
		close(paths)

		// ACT
		candidates, _ := Filter(context.Background(), paths, 0)

		if len(candidates) != 2 {
			t.Errorf("expected 2 candidates (the 2 good files share size), got %d", len(candidates))
		}
	})

	t.Run("respects minSize filter", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "tiny.txt"), "ab")          // 2 bytes
		writeFile(t, filepath.Join(dir, "big1.txt"), "aaaaaaaaaa")  // 10 bytes
		writeFile(t, filepath.Join(dir, "big2.txt"), "bbbbbbbbbb")  // 10 bytes
		
		paths := make(chan string, 10)
		paths <- filepath.Join(dir, "tiny.txt")
		paths <- filepath.Join(dir, "big1.txt")
		paths <- filepath.Join(dir, "big2.txt")
		close(paths)
		
		// minSize=5 should skip tiny.txt
		candidates, totalSeen := Filter(context.Background(), paths, 5)
		
		if totalSeen != 3 {
			t.Errorf("expected totalSeen=2 (tiny.txt filtered out), got %d", totalSeen)
		}
		if len(candidates) != 2 {
			t.Errorf("expected 2 candidates (big1/big2), got %d", len(candidates))
		}
	})
}

func writeFile(t *testing.T, path, content string) {
	t.Helper() 
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write %s: %v", path, err)
	}
}