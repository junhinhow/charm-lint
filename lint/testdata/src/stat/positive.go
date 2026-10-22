package stat

import "os"

// stdinMode descarta o erro de Stat e usa o FileInfo em seguida.
func stdinMode() os.FileMode {
	info, _ := os.Stdin.Stat()
	return info.Mode() // want `Stat falhar`
}

// stdoutSize descarta o erro de Stat antes de medir o tamanho.
func stdoutSize() int64 {
	info, _ := os.Stdout.Stat()
	return info.Size() // want `Stat falhar`
}

// stderrMode descarta o erro de Stat do stderr.
func stderrMode() os.FileMode {
	info, _ := os.Stderr.Stat()
	return info.Mode() // want `Stat falhar`
}
