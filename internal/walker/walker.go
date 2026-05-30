package walker

import (
	"context"
	"io/fs"
	"log"
	"path/filepath"
)

func Walk(ctx context.Context, root string, paths chan<- string) (int, error) {
	files := 0
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Printf("WARN: visit %s: %v", path, err)
			return nil
		}

		if d.IsDir() {
			return nil
		}

		select {
		case <-ctx.Done():
			return filepath.SkipAll
		case paths <- path:
			files++
			return nil
		}
	})
	return files, err
}