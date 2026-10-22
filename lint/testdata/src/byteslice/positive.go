package byteslice

import (
	"fmt"
	"strings"
)

// Model ilustra o fluxo típico de um componente Bubble Tea.
type Model struct{}

// Render corta o texto recebido por byte antes de devolvê-lo.
func (m Model) Render(s string, width int) string {
	return s[:width] // want `corta a string por byte`
}

func printSuffix(s string, n int) {
	fmt.Println(s[n:]) // want `corta a string por byte`
}

func concatPrefix(s string, n int) string {
	return "> " + s[:n] // want `corta a string por byte`
}

func upperThenPrint(s string, n int) {
	fmt.Print(strings.ToUpper(s[:n])) // want `corta a string por byte`
}

func sprintfLabel(s string, n int) string {
	label := s[:n] // want `corta a string por byte`
	return fmt.Sprintf("[%s]", label)
}

func emojiPrefix(s string, n int) string {
	return "🚀 " + s[:n] // want `corta a string por byte`
}
