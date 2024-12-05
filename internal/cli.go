package tsquery

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/urfave/cli/v3"
)

const (
	appName = "tsquery"
)

var (
	Version = "0.0.0"
	Time    string
)

func CLI(args []string, stdout io.Writer, stderr io.Writer) int {
	cmd := &cli.Command{
		Name:      appName,
		Usage:     "Learning things through spaced repetition.",
		Writer:    stdout,
		ErrWriter: stderr,
		UsageText: "<query> [file|directory]",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			var root, query string

			switch cmd.NArg() {
			case 0:
				return fmt.Errorf("no query specified")
			case 1:
				root = "."
				query = cmd.Args().First()
			case 2:
				query = cmd.Args().First()
				root = cmd.Args().Get(1)
			default:
				return fmt.Errorf("too many arguments")
			}

			return app(ctx, options{query: query, workers: 15, root: root, stdout: stdout})
		},
		Commands: []*cli.Command{
			{
				Name:    "version",
				Aliases: []string{"v"},
				Usage:   "Show version",
				Action: func(context.Context, *cli.Command) error {
					_, _ = fmt.Fprintf(stdout, "%s %s %s\n\n", appName, Version, Time)
					return nil
				},
			},
			{
				Name:  "languages",
				Usage: "Show supported languages",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					_, _ = fmt.Fprintf(stdout, "Languages: %s", strings.Join(languages, ", "))
					return nil
				},
			},
		},
	}

	err := cmd.Run(context.Background(), args)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "failed: %v\n", err)
		return -1
	}

	return 0
}
