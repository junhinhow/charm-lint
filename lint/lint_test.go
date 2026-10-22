package lint_test

import (
	"testing"

	"github.com/junhinhow/charm-lint/lint"
	"golang.org/x/tools/go/analysis/analysistest"
)

// TestCharmlint garante que o analisador agregado roda e relata as regras
// declaradas em Requires.
func TestCharmlint(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), lint.Charmlint, "aggregate")
}
