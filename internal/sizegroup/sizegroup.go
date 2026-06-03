package sizegroup

import (
	"context"
	"log"
	"os"
)

func Filter(ctx context.Context, paths <-chan string, minSize int64) ([]string, int) {
	sizeGroups := make(map[int64][]string)
	totalSeen := 0

	for {
		select {
		case <-ctx.Done():
			return collectCandidate(sizeGroups), totalSeen
		case path, ok := <-paths:
			if !ok {
				return collectCandidate(sizeGroups), totalSeen
			}

			
			fileInfo, err := os.Stat(path)
			if err != nil {
				log.Printf("WARN: failed stat: %v", err)
				continue
			}
			
			totalSeen++
			
			sz := fileInfo.Size()
			if sz < minSize {
				continue
			}
			sizeGroups[sz] = append(sizeGroups[sz], path)
		}
	}
}

func collectCandidate(sizeGroups map[int64][]string) []string {
	var candidates []string
	for _, paths := range sizeGroups {
		if len(paths) > 1 {
			candidates = append(candidates, paths...)
		}
	}

	return candidates
}