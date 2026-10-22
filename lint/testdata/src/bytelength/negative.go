package bytelength

import "strings"

// separator usa uma largura já calculada: não há medição de string.
func separator(width int) string {
	return strings.Repeat("-", width)
}

// padBytes mede um []byte, que não sofre do problema de largura de exibição.
func padBytes(b []byte) string {
	return strings.Repeat("x", len(b))
}

// padConstant usa um tamanho fixo.
func padConstant() string {
	return strings.Repeat("=", 10)
}

// repeatWord repete uma palavra por um número de vezes qualquer.
func repeatWord(word string, times int) string {
	return strings.Repeat(word, times)
}
