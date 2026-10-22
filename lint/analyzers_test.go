package lint_test

import (
	"testing"

	"github.com/junhinhow/charm-lint/lint"
)

func TestAnalyzers(t *testing.T) {
	analyzers := lint.Analyzers()
	if len(analyzers) != 6 {
		t.Fatalf("Analyzers() devolveu %d regras, esperado 6", len(analyzers))
	}

	seen := make(map[string]bool, len(analyzers))
	for _, analyzer := range analyzers {
		if analyzer.Name == "" {
			t.Error("regra sem nome")
		}
		if seen[analyzer.Name] {
			t.Errorf("regra duplicada: %s", analyzer.Name)
		}
		seen[analyzer.Name] = true
		if analyzer.Doc == "" {
			t.Errorf("regra %s sem documentação", analyzer.Name)
		}
		if analyzer.Run == nil {
			t.Errorf("regra %s sem Run", analyzer.Name)
		}
	}
}
