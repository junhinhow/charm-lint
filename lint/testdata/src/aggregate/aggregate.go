package aggregate

import (
	"os"
	"strings"
)

// pad calcula o padding por bytes.
func pad(s string) string {
	return s + strings.Repeat(" ", 10-len(s)) // want `padding calculado com len`
}

// statMode descarta o erro de Stat.
func statMode() os.FileMode {
	info, _ := os.Stdin.Stat()
	return info.Mode() // want `Stat falhar`
}
