# dedup-finder

A concurrent file deduplication tool written in Go. Finds duplicate files by content (SHA-256), regardless of filename or location.

## Features

- Concurrent processing using a fixed worker pool (8 goroutines)
- Streaming SHA-256 hashing — handles files of any size with constant memory
- Graceful cancellation via `context.Context` and `signal.Notify`
- Fault-tolerant — skips unreadable files without aborting the run
- Zero external dependencies

## Installation

Requires Go 1.21 or later.

```bash
git clone https://github.com/panditpankaj21/dedup-finder.git
cd dedup-finder
go build -o dedup ./cmd/dedup
```

## Usage

```bash
./dedup /path/to/directory
```

### Example

```
$ ./dedup ~/Pictures
Scanning ~/Pictures...
Took 4.2s

Scanned 12,847 files

Group 1 (hash 6ac3a507...) — 3 files:
 ~/Pictures/IMG_1234.jpg
 ~/Pictures/backup/old_photo.jpg
 ~/Pictures/2024/copy.jpg

Found 1,243 duplicate group(s).
```

Press `Ctrl+C` at any time for graceful shutdown with partial results.

## How It Works

1. A walker traverses the directory tree via `filepath.WalkDir`
2. File paths flow through a buffered channel to 8 worker goroutines
3. Each worker streams its file through SHA-256 using `io.Copy`
4. Results aggregate into a `map[string][]string` protected by `sync.Mutex`
5. `context.Context` propagates cancellation to all goroutines on `Ctrl+C`

## Tech Stack

Go standard library only: `crypto/sha256`, `context`, `io`, `os`, `path/filepath`, `sync`, `os/signal`.

## License

MIT