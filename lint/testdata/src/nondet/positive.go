package nondet

import (
	"fmt"
	"strings"
)

// Model guarda itens em um map, como muitos componentes reais.
type Model struct {
	items map[string]int
}

// Render imprime o map diretamente: a ordem muda a cada execução.
func (m Model) Render() string {
	var b strings.Builder
	for name, count := range m.items { // want `ordem aleatória`
		fmt.Fprintf(&b, "%s: %d\n", name, count)
	}
	return b.String()
}

// View concatena as chaves na ordem aleatória do map.
func (m Model) View() string {
	out := ""
	for name := range m.items { // want `ordem aleatória`
		out += name + "\n"
	}
	return out
}

// String escreve as chaves direto no buffer.
func (m Model) String() string {
	var b strings.Builder
	for name := range m.items { // want `ordem aleatória`
		b.WriteString(name)
	}
	return b.String()
}
