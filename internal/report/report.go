package report

import "fmt"

func Print(hashes map[string][]string) int {
    groups := 0
    for h, paths := range hashes {
        if len(paths) <= 1 {
            continue
        }
        groups++
        fmt.Printf("\nGroup %d (hash %s...) — %d files:\n", groups, h[:8], len(paths))
        for _, p := range paths {
            fmt.Printf(" %s\n", p)
        }
    }
    return groups
}