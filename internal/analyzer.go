package tsquery

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/go-enry/go-enry/v2"
	sitter "github.com/smacker/go-tree-sitter"
)

type capture struct {
	content string
	row     uint32
	column  uint32
}

func analyzeFile(ctx context.Context, path, query string) ([]capture, error) {
	abspath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("could not resolve absolute path %s: %w", path, err)
	}

	contents, err := os.ReadFile(abspath)
	if err != nil {
		return nil, fmt.Errorf("failed read file %s: %w", abspath, err)
	}

	enryLanguage := enry.GetLanguage(abspath, contents)
	if enryLanguage == "" {
		return nil, ErrLangNotDetected
	}

	tsLanguage, err := languageFromEnry(enryLanguage)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
	}

	parser := sitter.NewParser()
	parser.SetLanguage(tsLanguage)

	tree, err := parser.ParseCtx(ctx, nil, contents)
	if err != nil {
		return nil, err
	}
	root := tree.RootNode()

	q, err := sitter.NewQuery([]byte(query), tsLanguage)
	if err != nil {
		return nil, fmt.Errorf("failed creating query: %w", err)
	}
	defer q.Close()

	qc := sitter.NewQueryCursor()
	defer qc.Close()

	qc.Exec(q, root)

	var captures []capture
	for {
		m, ok := qc.NextMatch()
		if !ok {
			break
		}

		for _, c := range m.Captures {
			captures = append(captures,
				capture{
					content: c.Node.Content(contents),
					row:     c.Node.StartPoint().Row + 1,
					column:  c.Node.StartPoint().Column + 1,
				},
			)
		}
	}

	return captures, nil
}

type result struct {
	path     string
	captures []capture
	err      error
}

func analyzers(ctx context.Context, query string, paths <-chan string, results chan<- result) {
	for path := range paths {
		captures, err := analyzeFile(ctx, path, query)
		select {
		case results <- result{path, captures, err}:
		case <-ctx.Done():
			return
		}
	}
}
