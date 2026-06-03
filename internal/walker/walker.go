package walker

import (
	"context"
	"io/fs"
	"log"
	"path/filepath"
)

func Walk(ctx context.Context, root string, ignorePatterns []string, paths chan<- string) (int, error) {
	files := 0
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Printf("WARN: visit %s: %v", path, err)
			return nil
		}

		name := d.Name()
        for _, pattern := range ignorePatterns {
            matched, _ := filepath.Match(pattern, name)
            if matched {
                if d.IsDir() {
                    return filepath.SkipDir  // skip entire directory
                }
                return nil  // skip this file
            }
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