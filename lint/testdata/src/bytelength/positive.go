package bytelength

import (
	"strings"
	"unicode/utf8"
)

// padLen mede o texto por bytes: com acento o padding sai maior que o
// esperado e a coluna desalinha.
func padLen(label string) string {
	return label + strings.Repeat(" ", 20-len(label)) // want `padding calculado com len`
}

// padRunes conta runes, mas um emoji ocupa duas células e o padding sai
// menor que o esperado.
func padRunes(label string) string {
	return label + strings.Repeat("─", utf8.RuneCountInString(label)) // want `RuneCountInString`
}

// padEmoji monta um título com emoji e calcula o preenchimento em bytes.
func padEmoji(label string) string {
	titulo := "🚀 " + label
	return titulo + strings.Repeat("·", 24-len(titulo)) // want `padding calculado com len`
}
