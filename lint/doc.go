// Package lint fornece analisadores estáticos para armadilhas comuns em
// interfaces de terminal (TUI) construídas com o Charm Suite.
//
// As regras cobrem problemas que só aparecem em tempo de execução, como
// padding calculado em bytes, corte de strings no meio de um rune, iteração
// sobre map em funções de renderização e erros de comandos externos
// descartados.
//
// Cada regra é exportada como um *analysis.Analyzer independente. O
// analisador [Charmlint] agrega todas elas e é o ponto de entrada usado pelo
// binário cmd/charm-lint.
package lint
