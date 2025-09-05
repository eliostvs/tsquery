package tsquery

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-enry/go-enry/v2"
	sitter "github.com/smacker/go-tree-sitter"
)

type Capture struct {
	Content string
	Row     uint32
	Column  uint32
}

type Result struct {
	Filepath string
	Captures []Capture
}

func analyzeFile(ctx context.Context, path, query string) (*Result, error) {
	abspath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("could not resolve absolute path %s: %w", path, err)
	}

	contents, err := os.ReadFile(abspath)
	if err != nil {
		return nil, fmt.Errorf("read file %s: %w", abspath, err)
	}

	enryLanguage := enry.GetLanguage(abspath, contents)
	if enryLanguage == "" {
		return nil, ErrLangNotDetected
	}

	sitterLanguage, err := languageFromEnry(enryLanguage)
	if errors.Is(err, ErrLangNotSupported) {
		return nil, err
	}

	parser := sitter.NewParser()
	parser.SetLanguage(sitterLanguage)

	tree, err := parser.ParseCtx(ctx, nil, contents)
	if err != nil {
		return nil, fmt.Errorf("parsing query: %w", err)
	}
	root := tree.RootNode()

	q, err := sitter.NewQuery([]byte(query), sitterLanguage)
	if err != nil {
		return nil, fmt.Errorf("creating query: %w", err)
	}
	defer q.Close()

	qc := sitter.NewQueryCursor()
	defer qc.Close()

	qc.Exec(q, root)

	var captures []Capture
	for {
		m, ok := qc.NextMatch()
		if !ok {
			break
		}

		for _, c := range m.Captures {
			captures = append(captures,
				Capture{
					Content: c.Node.Content(contents),
					Row:     c.Node.StartPoint().Row + 1,
					Column:  c.Node.StartPoint().Column + 1,
				},
			)
		}
	}

	return &Result{Filepath: abspath, Captures: captures}, nil
}

func analyzeFilePipeline(ctx context.Context, query string, paths <-chan string, errs chan<- error) <-chan *Result {
	results := make(chan *Result)

	go func() {
		defer close(results)
		for path := range paths {
			select {
			case <-ctx.Done():
				return
			default:
			}

			result, err := analyzeFile(ctx, path, query)
			if err != nil {
				select {
				case errs <- err:
					continue
				case <-ctx.Done():
					return
				}
			}

			select {
			case results <- result:
			case <-ctx.Done():
				return
			}
		}
	}()

	return results
}
