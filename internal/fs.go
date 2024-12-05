package tsquery

import (
	"context"
	"errors"
	"io/fs"
	"path/filepath"
	"strings"
)

var ErrWalkCancelled = errors.New("walk cancelled")

func walkDir(ctx context.Context, root string) (<-chan string, <-chan error) {
	paths := make(chan string)
	errc := make(chan error, 1)

	go func() {
		defer close(paths)

		errc <- filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if d != nil && d.IsDir() && strings.HasPrefix(d.Name(), ".") {
				return fs.SkipDir
			}

			if d == nil && d.IsDir() {
				return nil
			}

			select {
			case paths <- path:
			case <-ctx.Done():
				return ErrWalkCancelled
			}

			return nil
		})
	}()

	return paths, errc
}
