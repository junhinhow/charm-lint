// Command charm-lint roda os analisadores estáticos do Charm Suite.
//
// Uso:
//
//	charm-lint ./...
//
// O binário também funciona como vettool:
//
//	go vet -vettool=$(which charm-lint) ./...
package main

import (
	"github.com/junhinhow/charm-lint/lint"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(lint.Charmlint)
}
