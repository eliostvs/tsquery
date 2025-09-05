package tsquery

import (
	"context"
	"io/fs"
	"path/filepath"
	"strings"
)

func walkDirPipeline(ctx context.Context, root string, errc chan<- error) <-chan string {
	paths := make(chan string)

	go func() {
		defer close(paths)

		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if d != nil && d.IsDir() && strings.HasPrefix(d.Name(), ".") {
				return fs.SkipDir
			}

			if d != nil && d.IsDir() {
				return nil
			}

			select {
			case paths <- path:
			case <-ctx.Done():
				return ctx.Err()
			}

			return nil
		})
		if err != nil {
			select {
			case errc <- err:
			case <-ctx.Done():
			}
		}
	}()

	return paths
}
