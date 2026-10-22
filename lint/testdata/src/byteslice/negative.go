package byteslice

import "strings"

// fixedPrefix usa limite constante: não há risco de cair no meio de um rune
// em tempo de execução.
func fixedPrefix(s string) string {
	return s[:3]
}

// bytesPrefix fatia []byte, não string.
func bytesPrefix(b []byte, n int) []byte {
	return b[:n]
}

// prefixCheck usa o slice para comparação, não para exibição.
func prefixCheck(s string, n int) bool {
	return strings.HasPrefix(s[:n], ">")
}

// useAsKey usa o slice como chave de map, sem renderizar.
func useAsKey(s string, n int, m map[string]int) {
	m[s[:n]] = 1
}

// measure usa o slice apenas para medir o tamanho.
func measure(s string, n int) int {
	part := s[:n]
	return len(part)
}

// nestedLength mede o slice dentro de outra expressão.
func nestedLength(s string, n int, out *int) {
	*out = 2 * len(s[:n])
}
