package lint_test

import (
	"testing"

	"github.com/junhinhow/charm-lint/lint"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestSilentExec(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), lint.SilentExec, "silentexec")
}
