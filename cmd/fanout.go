package cmd

import (
	"errors"
	"fmt"
	"os"
	"sync"
)

// runConcurrent runs each fn concurrently, waits for all to finish, and
// returns all non-nil errors joined via errors.Join.
func runConcurrent(fns ...func() error) error {
	errs := make([]error, len(fns))
	var wg sync.WaitGroup
	for i, fn := range fns {
		wg.Add(1)
		go func(i int, fn func() error) {
			defer wg.Done()
			errs[i] = fn()
		}(i, fn)
	}
	wg.Wait()
	return errors.Join(errs...)
}

// pollAndDiff fetches items, records their IDs in the returned seen set, and
// calls handle for each item not in the provided seen set (skipped during the
// seeding pass). Returns the updated seen set; returns the original on error.
// filter may be nil to accept all items.
func pollAndDiff[T any](
	ts string,
	fetch func() ([]T, error),
	getID func(T) string,
	filter func(T) bool,
	handle func(T),
	seen map[string]struct{},
	seeding bool,
	resourceName string,
) map[string]struct{} {
	items, err := fetch()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[%s] Error listing %s: %v\n", ts, resourceName, err)
		return seen
	}
	current := make(map[string]struct{})
	for _, item := range items {
		if filter != nil && !filter(item) {
			continue
		}
		id := getID(item)
		current[id] = struct{}{}
		if _, ok := seen[id]; ok || seeding {
			continue
		}
		handle(item)
	}
	return current
}
