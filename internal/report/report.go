package report

import "fmt"

func Print(hashes map[string][]string) int {
    groups := 0
    for h, paths := range hashes {
        if len(paths) <= 1 {
            continue
        }
        groups++
        short := h
        if len(h) > 8 {
            short = h[:8]
        }
        fmt.Printf("\nGroup %d (hash %s...) — %d files:\n", groups, short, len(paths))
        for _, p := range paths {
            fmt.Printf(" %s\n", p)
        }
    }
    return groups
}