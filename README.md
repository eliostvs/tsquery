# tsquery

`tsquery` is a CLI for executing [Tree-sitter queries](https://tree-sitter.github.io/tree-sitter/using-parsers#query-syntax) on source code files.
It uses [`enry`](https://github.com/go-enry/go-enry) to detect a language and apply the right Tree-sitter [parser](https://tree-sitter.github.io/tree-sitter/#available-parsers).
The default output includes a list of line number locations where there's a query match, followed by a snippet of the matching code.
