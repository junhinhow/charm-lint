package lint

import (
	"fmt"
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// SilentExec detecta comandos externos cujo erro ou cuja saída de erro é
// descartada.
var SilentExec = &analysis.Analyzer{
	Name: "silentexec",
	Doc: `Detecta exec.Command(...).Run/Output/CombinedOutput com o erro descartado ou sem cmd.Stderr configurado.

Quando o erro é ignorado ou o stderr não é ligado a lugar nenhum, uma falha do comando passa despercebida e o usuário vê a interface "funcionando" sem o resultado esperado. Trate o erro retornado e defina cmd.Stderr = os.Stderr.`,
	Run: runSilentExec,
}

// runSilentExec procura chamadas de execução de comando que escondem falhas.
func runSilentExec(pass *analysis.Pass) (any, error) {
	commandVars := execCommandVars(pass)
	for _, file := range pass.Files {
		parents := parentsOf(file)
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := ast.Unparen(call.Fun).(*ast.SelectorExpr)
			if !ok {
				return true
			}
			switch sel.Sel.Name {
			case "Run", "Output", "CombinedOutput":
			default:
				return true
			}
			if !isExecCommandReceiver(pass, sel.X, commandVars) {
				return true
			}

			discarded := isResultDiscarded(parents[call], errorIndex(sel.Sel.Name))
			fn := enclosingFunc(pass, call.Pos())
			hasStderr := fn != nil && assignsStderr(fn.Body)
			if !discarded && hasStderr {
				return true
			}

			text := types.ExprString(call)
			var msg string
			switch {
			case discarded && !hasStderr:
				msg = fmt.Sprintf("%s: erro descartado e cmd.Stderr não configurado — a falha do comando fica invisível; trate o erro e defina cmd.Stderr = os.Stderr", text)
			case discarded:
				msg = fmt.Sprintf("%s: erro descartado — a falha do comando fica invisível; trate o erro em vez de ignorá-lo", text)
			default:
				msg = fmt.Sprintf("%s: cmd.Stderr não configurado — a mensagem de erro do comando se perde; defina cmd.Stderr = os.Stderr e trate o erro", text)
			}
			pass.Report(analysis.Diagnostic{
				Pos:      call.Pos(),
				End:      call.End(),
				Category: ruleSilentExec,
				Message:  msg,
			})
			return true
		})
	}
	return nil, nil
}

// execCommandVars devolve as variáveis atribuídas a partir de exec.Command,
// para reconhecer o padrão "cmd := exec.Command(...); cmd.Run()".
func execCommandVars(pass *analysis.Pass) map[types.Object]bool {
	vars := make(map[types.Object]bool)
	record := func(expr ast.Expr, target ast.Expr) {
		if !isExecCommandCall(pass, expr) {
			return
		}
		id, ok := ast.Unparen(target).(*ast.Ident)
		if !ok || id.Name == "_" {
			return
		}
		if obj := pass.TypesInfo.ObjectOf(id); obj != nil {
			vars[obj] = true
		}
	}
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.AssignStmt:
				for i, rhs := range node.Rhs {
					if i < len(node.Lhs) {
						record(rhs, node.Lhs[i])
					}
				}
			case *ast.ValueSpec:
				for i, value := range node.Values {
					if i < len(node.Names) {
						record(value, node.Names[i])
					}
				}
			}
			return true
		})
	}
	return vars
}

// isExecCommandCall informa se expr é uma chamada a exec.Command.
func isExecCommandCall(pass *analysis.Pass, expr ast.Expr) bool {
	call, ok := ast.Unparen(expr).(*ast.CallExpr)
	return ok && isPkgFunc(pass, call, "os/exec", "Command")
}

// isExecCommandReceiver informa se x é um comando criado por exec.Command,
// seja na forma encadeada, seja por uma variável.
func isExecCommandReceiver(pass *analysis.Pass, x ast.Expr, commandVars map[types.Object]bool) bool {
	if isExecCommandCall(pass, x) {
		return true
	}
	id, ok := ast.Unparen(x).(*ast.Ident)
	if !ok {
		return false
	}
	return commandVars[pass.TypesInfo.ObjectOf(id)]
}

// isResultDiscarded informa se o resultado da chamada é jogado fora: usada
// como expressão solta ou com o erro atribuído a "_".
//
// errorIndex é a posição do erro entre os resultados da chamada (0 para Run,
// 1 para Output e CombinedOutput).
func isResultDiscarded(parent ast.Node, errorIndex int) bool {
	switch node := parent.(type) {
	case *ast.ExprStmt:
		return true
	case *ast.AssignStmt:
		if errorIndex >= len(node.Lhs) {
			return false
		}
		id, ok := ast.Unparen(node.Lhs[errorIndex]).(*ast.Ident)
		return ok && id.Name == "_"
	}
	return false
}

// errorIndex devolve a posição do erro entre os resultados do método.
func errorIndex(method string) int {
	if method == "Run" {
		return 0
	}
	return 1
}

// assignsStderr informa se o corpo da função atribui algo a um campo Stderr.
func assignsStderr(body *ast.BlockStmt) bool {
	found := false
	ast.Inspect(body, func(n ast.Node) bool {
		if found {
			return false
		}
		assign, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}
		for _, lhs := range assign.Lhs {
			sel, ok := ast.Unparen(lhs).(*ast.SelectorExpr)
			if ok && sel.Sel.Name == "Stderr" {
				found = true
				return false
			}
		}
		return true
	})
	return found
}
