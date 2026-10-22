package silentexec

import (
	"os"
	"os/exec"
)

// runOK propaga o erro e liga o stderr do comando.
func runOK() error {
	cmd := exec.Command("ls")
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// outputOK propaga o erro e liga o stderr do comando.
func outputOK() ([]byte, error) {
	cmd := exec.Command("ls")
	cmd.Stderr = os.Stderr
	return cmd.Output()
}

// combinedOK propaga o erro e liga o stderr do comando.
func combinedOK() ([]byte, error) {
	cmd := exec.Command("ls")
	cmd.Stderr = os.Stderr
	return cmd.CombinedOutput()
}

// outputErroTratado descarta a saída, mas trata o erro.
func outputErroTratado() error {
	cmd := exec.Command("date")
	cmd.Stderr = os.Stderr
	_, err := cmd.Output()
	return err
}

// startOnly usa Start, que não faz parte da regra.
func startOnly() error {
	cmd := exec.Command("ls")
	return cmd.Start()
}

// fakeCmd não é um comando de exec.
type fakeCmd struct{}

// Run simula a interface de um comando qualquer.
func (fakeCmd) Run() error { return nil }

// notExec chama Run em um tipo que não vem de exec.Command.
func notExec() error {
	var cmd fakeCmd
	return cmd.Run()
}

// otherStderr atribui Stderr de outro comando na mesma função: a checagem é
// por função, então esta chamada não é sinalizada.
func otherStderr() error {
	other := exec.Command("true")
	other.Stderr = os.Stderr
	return exec.Command("ls").Run()
}
