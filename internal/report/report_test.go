package report

import "testing"

func TestPrint(t *testing.T) {
	t.Run("counts duplicate groups correctly", func(t *testing.T) {
		// ARRANGE: 3 unique hashes, 2 of them have duplicates
		hashes := map[string][]string{
			"hash_a": {"file1.txt", "file2.txt"},          // duplicate group (2 files)
			"hash_b": {"file3.txt", "file4.txt", "file5.txt"}, // duplicate group (3 files)
			"hash_c": {"file6.txt"},                       // NOT a duplicate (only 1 file)
		}

		// ACT
		got := Print(hashes)

		// ASSERT: should count only groups where len > 1, so 2 groups
		want := 2
		if got != want {
			t.Errorf("expected %d duplicate groups, got %d", want, got)
		}
	})

	t.Run("empty map produces zero groups", func(t *testing.T) {
		// Edge case: empty input
		hashes := map[string][]string{}
		got := Print(hashes)
		if got != 0 {
			t.Errorf("expected 0 groups for empty map, got %d", got)
		}
	})

	t.Run("all singletons produce zero groups", func(t *testing.T) {
		// All unique files = no duplicates
		hashes := map[string][]string{
			"hash_a": {"file1.txt"},
			"hash_b": {"file2.txt"},
			"hash_c": {"file3.txt"},
		}
		got := Print(hashes)
		if got != 0 {
			t.Errorf("expected 0 groups (all singletons), got %d", got)
		}
	})
}