package tsquery

import (
	"context"
	"fmt"
	"io"
	"sync"
)

type options struct {
	query   string
	workers int
	root    string
	stdout  io.Writer
}

func app(ctx context.Context, opts options) error {
	ctx, cancel := context.WithCancel(ctx)

	defer func() {
		cancel()
	}()

	paths, errc := walkDir(ctx, opts.root)
	results := make(chan result)
	var wg sync.WaitGroup
	wg.Add(opts.workers)

	for i := 0; i < opts.workers; i++ {
		go func() {
			analyzers(ctx, opts.query, paths, results)
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for result := range results {
		if result.err != nil {
			fmt.Fprintf(opts.stdout, "%s: %s", result.path, result.err.Error())
		}

		for _, capture := range result.captures {
			fmt.Fprintf(opts.stdout, "%s:%d:%d\n", result.path, capture.row, capture.column)
			fmt.Fprintln(opts.stdout, capture.content)
		}
	}

	if err := <-errc; err != nil {
		return err
	}

	return nil
}
