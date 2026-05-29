package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

func main(){
	if len(os.Args) < 2 {
		log.Fatal("usage: dedup <directory>")
	}
	root := os.Args[1]
	hashes := make(map[string][]string)

	log.Printf("Scanning %s...", root)
	files := 0
	start := time.Now()

	var wg sync.WaitGroup
	var mu sync.Mutex

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Printf("WARN: could not read %s: %v", path, err)
			return nil
		}

		if d.IsDir() {
			// if it is directory skip
			return nil
		}

		files++
		wg.Add(1)
		go func(path string) { 
			defer wg.Done()

			data, err := os.ReadFile(path)
			if err != nil {
				log.Printf("WARN: could not read %s: %v", path, err)
				return
			}

			hash := sha256.Sum256(data) // returns [32]byte fixed-size array
			// hex.EncodeToString gives you the readable hash like "2cf24dba...".
			hashStr := hex.EncodeToString(hash[:]) // converts to 64-char hex string

			mu.Lock()
			hashes[hashStr] = append(hashes[hashStr], path)
			mu.Unlock()
			
		}(path)

		return nil

	})

	if err != nil {
		fmt.Printf("Error while traversing the path %q: %v", root, err)
	}

	wg.Wait()

	elapsed := time.Since(start)
	fmt.Printf("Took %v\n", elapsed)

	fmt.Println()
	fmt.Printf("Scanned %d files \n", files)

	duplicateGroups := 0

	for h, paths := range hashes {
		if len(paths) > 1 {
			duplicateGroups++
			fmt.Printf("\nGroup %d (hash %s...) — %d files:\n", duplicateGroups, h[:8], len(paths))
			for _, p := range paths {
				fmt.Printf(" %s\n", p)
			}
		}
	}

	if duplicateGroups == 0 {
		fmt.Println("\nNo duplicates found.")
	} else {
		fmt.Printf("\nFound %d duplicate group(s).\n", duplicateGroups)
	}
}