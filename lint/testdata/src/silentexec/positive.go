package silentexec

import "os/exec"

// runDiscarded joga fora o erro de Run com "_ =".
func runDiscarded() {
	_ = exec.Command("git", "status").Run() // want `erro descartado`
}

// outputDiscarded ignora o erro de Output.
func outputDiscarded() []byte {
	out, _ := exec.Command("date").Output() // want `erro descartado`
	return out
}

// runWithoutStderr não liga o stderr do comando em lugar nenhum.
func runWithoutStderr() error {
	cmd := exec.Command("ls")
	return cmd.Run() // want `cmd.Stderr não configurado`
}

// combinedWithoutStderr também deixa o stderr sem destino explícito.
func combinedWithoutStderr() ([]byte, error) {
	cmd := exec.Command("ls")
	return cmd.CombinedOutput() // want `cmd.Stderr não configurado`
}
