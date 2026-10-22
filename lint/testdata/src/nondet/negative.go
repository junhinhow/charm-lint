package nondet

import (
	"fmt"
	"sort"
	"strings"
)

// SortedModel sempre ordena as chaves antes de renderizar.
type SortedModel struct {
	items map[string]int
}

// String coleta as chaves, ordena e só depois monta a saída.
func (m SortedModel) String() string {
	names := make([]string, 0, len(m.items))
	for name := range m.items { // apenas coleta: a ordenação vem depois
		names = append(names, name)
	}
	sort.Strings(names)

	var b strings.Builder
	for _, name := range names {
		fmt.Fprintf(&b, "%s: %d\n", name, m.items[name])
	}
	return b.String()
}

// RenderSorted delega a ordenação para um helper.
func (m SortedModel) RenderSorted() string {
	var b strings.Builder
	for _, name := range sortedNames(m.items) {
		fmt.Fprintf(&b, "%s: %d\n", name, m.items[name])
	}
	return b.String()
}

// helper não é uma função de renderização.
func (m SortedModel) helper() string {
	for name := range m.items {
		_ = name
	}
	return ""
}

// RenderStatic não itera sobre map.
func (m SortedModel) RenderStatic() string {
	return "ok"
}

// sortedNames é o helper que ordena as chaves.
func sortedNames(items map[string]int) []string {
	names := make([]string, 0, len(items))
	for name := range items {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
