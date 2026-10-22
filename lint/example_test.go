package lint_test

import (
	"fmt"

	"github.com/junhinhow/charm-lint/lint"
)

// ExampleAnalyzers mostra as regras na ordem em que são executadas.
func ExampleAnalyzers() {
	for _, analyzer := range lint.Analyzers() {
		fmt.Println(analyzer.Name)
	}
	// Output:
	// bytelengthpadding
	// byteslice
	// rangeindexarith
	// nondeterministicrender
	// silentexec
	// discardedstaterror
}
