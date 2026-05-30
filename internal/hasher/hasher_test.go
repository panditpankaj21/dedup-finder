package hasher

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFile(t *testing.T) {
	t.Run("known content gives known hash", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "hello.txt")
		if err := os.WriteFile(path, []byte("hello"), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}

		got, err := File(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// SHA-256 of "hello" is a known value
		want := "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
		if got != want {
			t.Errorf("got %s, want %s", got, want)
		}
	})

	t.Run("empty file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "empty.txt")
		if err := os.WriteFile(path, []byte(""), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}

		got, err := File(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// SHA-256 of empty input is a well-known value
		want := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
		if got != want {
			t.Errorf("got %s, want %s", got, want)
		}
	})

	t.Run("same content different files gives same hash", func(t *testing.T) {
		dir := t.TempDir()
		pathA := filepath.Join(dir, "a.txt")
		pathB := filepath.Join(dir, "b.txt")

		content := []byte("identical bytes")
		os.WriteFile(pathA, content, 0644)
		os.WriteFile(pathB, content, 0644)

		hashA, _ := File(pathA)
		hashB, _ := File(pathB)

		if hashA != hashB {
			t.Errorf("expected same hash, got %s and %s", hashA, hashB)
		}
	})

	t.Run("nonexistent file returns error", func(t *testing.T) {
		_, err := File("/this/path/does/not/exist")
		if err == nil {
			t.Error("expected error for nonexistent file, got nil")
		}
	})
}