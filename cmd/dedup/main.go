package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const numWorker = 8

func main(){
	if len(os.Args) < 2 {
		log.Fatal("usage: dedup <directory>")
	}
	root := os.Args[1]
	
	log.Printf("Scanning %s...", root)
	start := time.Now()
	
	paths := make(chan string, 100)
	var wg sync.WaitGroup
	var mu sync.Mutex
	hashes := make(map[string][]string)
	files := 0

	//start 8 wokers
	for i:=0; i<numWorker; i++ {
		wg.Add(1)
		go func(){
			defer wg.Done()
			for path := range paths {
				// read -> hash -> lock -> append -> unlock
				hashStr, err := hashFile(path)
				if err != nil {
					log.Printf("WARN: could not hash %s: %v", path, err)
					continue
				}

				mu.Lock()
				hashes[hashStr] = append(hashes[hashStr], path)
				mu.Unlock()
			}
		}()
	}

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
		paths <- path
		return nil
	})
	close(paths)

	if err != nil {
		log.Printf("Error while traversing the path %q: %v", root, err)
	}

	wg.Wait()

	elapsed := time.Since(start)
	fmt.Printf("Took %v\n", elapsed)

	fmt.Println()
	fmt.Printf("Scanned %d files \n", files)

	duplicateGroups := 0

	for h, pathList := range hashes {
		if len(pathList) > 1 {
			duplicateGroups++
			fmt.Printf("\nGroup %d (hash %s...) — %d files:\n", duplicateGroups, h[:8], len(pathList))
			for _, p := range pathList {
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


func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()

	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	// Loop:
	// 	read 32KB from f into buffer
	// 	if EOF: break
	// 	write buffer to h
	// 	discard buffer

	sum := h.Sum(nil) // []byte (32 bytes)
	hashStr := hex.EncodeToString(sum)

	// hash := sha256.Sum256(data) // returns [32]byte fixed-size array
	// // hex.EncodeToString gives you the readable hash like "2cf24dba...".
	// hashStr := hex.EncodeToString(hash[:]) // converts to 64-char hex string

	return hashStr, nil
}