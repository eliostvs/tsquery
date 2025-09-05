package tsquery

import (
	"context"
	"fmt"
	"io"
	"time"
)

type options struct {
	query   string
	workers int
	root    string
	stdout  io.Writer
	timeout time.Duration
}

func app(ctx context.Context, opts options) error {
	errs := make(chan error, 1)

	if opts.timeout == 0 {
		opts.timeout = time.Minute
	}

	ctxt, cancel := context.WithTimeout(ctx, opts.timeout)
	defer cancel()

	paths := walkDirPipeline(ctxt, opts.root, errs)

	workers := fanOut(ctxt, paths, opts.workers)

	var channels []<-chan *Result
	for _, ch := range workers {
		channels = append(channels, analyzeFilePipeline(ctxt, opts.query, ch, errs))
	}

	results := fanIn(ctxt, channels...)

	printResults(opts.stdout, results)

	if err := <-errs; err != nil {
		return err
	}

	return nil
}

func printResults(stdout io.Writer, results <-chan *Result) {
	for result := range results {
		for _, capture := range result.Captures {
			fmt.Fprintf(stdout, "%s:%d:%d\n", result.Filepath, capture.Row, capture.Column)
			fmt.Fprintln(stdout, capture.Content)
		}
	}
}
