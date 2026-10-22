package lint

import (
	"fmt"

	"golang.org/x/tools/go/analysis"
)

// Charmlint agrega todas as regras do charm-lint em um único analisador.
//
// É o analisador usado pelo binário cmd/charm-lint: ele executa cada regra
// sobre o mesmo [analysis.Pass] e relata os diagnósticos com a categoria da
// regra. A agregação é feita em Run, e não em Requires, porque o driver de
// singlechecker imprime apenas os diagnósticos do analisador raiz.
var Charmlint = &analysis.Analyzer{
	Name: "charmlint",
	Doc: `Detecta armadilhas comuns de TUI no estilo Charm: padding medido em bytes, corte de string no meio de um rune, índice de range usado como índice de rune, iteração de map em renderização e erros de comandos externos descartados.

Rodar "charm-lint ./..." aplica todas as regras do pacote lint. Cada regra também está disponível isoladamente para uso com unitchecker ou go vet -vettool.`,
	Run: func(pass *analysis.Pass) (any, error) {
		for _, analyzer := range Analyzers() {
			if _, err := analyzer.Run(pass); err != nil {
				return nil, fmt.Errorf("%s: %w", analyzer.Name, err)
			}
		}
		return nil, nil
	},
}

// Analyzers devolve todas as regras do charm-lint na ordem em que são
// executadas.
func Analyzers() []*analysis.Analyzer {
	return []*analysis.Analyzer{
		ByteLengthPadding,
		ByteSlice,
		RangeIndexArith,
		NondeterministicRender,
		SilentExec,
		DiscardedStatError,
	}
}
