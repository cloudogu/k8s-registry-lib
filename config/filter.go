package config

import "strings"

// WatchFilter can be applied to the result of a comparison of two configurations.
// It is a predicate function that should to return true in case the filter matches.
type WatchFilter func([]DiffResult) bool

// KeyFilter is a WatchFilter to watch for changes for a single Key.
func KeyFilter(k Key) WatchFilter {
	k = sanitizeKey(k)

	return func(diffs []DiffResult) bool {
		for _, diff := range diffs {
			if diff.Key == k {
				return true
			}
		}

		return false
	}
}

// DirectoryFilter is a WatchFilter to watch for changes for a directory.
// This can encompass multiple keys with the same prefix.
// For example:
// Filter with Key "key1/" watch for changes for the keys:
// key1/key2/key3
// key1/key2/key3
// ...
func DirectoryFilter(directory Key) WatchFilter {
	directory = sanitizeKey(directory)
	if directory != "" && !strings.HasSuffix(directory.String(), keySeparator) {
		directory = directory + keySeparator
	}

	return func(diffs []DiffResult) bool {
		for _, diff := range diffs {
			if strings.HasPrefix(diff.Key.String(), directory.String()) {
				return true
			}
		}

		return false
	}
}
