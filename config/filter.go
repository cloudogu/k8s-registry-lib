package config

import "strings"

type WatchFilter func([]DiffResult) bool

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
