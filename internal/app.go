package tsquery

import (
	"context"
	"errors"
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
		opts.timeout = 5 * time.Minute
	}

	ctxt, cancel := context.WithTimeout(ctx, opts.timeout)
	defer cancel()

	paths := walkDirPipeline(ctxt, opts.root, errs)

	workers := fanOut(paths, opts.workers)

	var channels []<-chan *Result
	for _, ch := range workers {
		channels = append(channels, analyzeFilePipeline(ctxt, opts.query, ch, errs))
	}

	results := fanIn(channels...)

	printResults(opts.stdout, results)

	close(errs)
	var errorList []error
	for err := range errs {
		errorList = append(errorList, err)
	}

	if len(errorList) > 0 {
		return errors.Join(errorList...)
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
