package stat

import "os"

// checkedMode verifica o erro antes de usar o FileInfo.
func checkedMode() (os.FileMode, error) {
	info, err := os.Stdin.Stat()
	if err != nil {
		return 0, err
	}
	return info.Mode(), nil
}

// nameOnly usa apenas Name, que não é o foco da regra.
func nameOnly() string {
	info, _ := os.Stdin.Stat()
	return info.Name()
}

// fakeFile simula um objeto com Stat.
type fakeFile struct{}

// Stat devolve um FileInfo nil sem erro.
func (fakeFile) Stat() (os.FileInfo, error) { return nil, nil }

// customStat chama Stat em um tipo que não é os.Stdin/Stdout/Stderr.
func customStat() os.FileMode {
	info, _ := fakeFile{}.Stat()
	return info.Mode()
}

// regularFile chama Stat em um arquivo aberto, fora do escopo da regra.
func regularFile() os.FileMode {
	f, err := os.Open("go.mod")
	if err != nil {
		return 0
	}
	defer f.Close()

	info, _ := f.Stat()
	return info.Mode()
}
