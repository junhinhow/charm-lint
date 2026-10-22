package lint

import (
	"fmt"
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// DiscardedStatError detecta o erro de os.Stdin/Stdout/Stderr.Stat
// descartado antes de usar o FileInfo.
var DiscardedStatError = &analysis.Analyzer{
	Name: "discardedstaterror",
	Doc: `Detecta "fi, _ := os.Stdin.Stat()" (ou Stdout/Stderr) seguido de fi.Mode()/fi.Size().

Se Stat falhar, o FileInfo devolvido é nil e o uso seguinte provoca panic. Verifique o erro antes de usar Mode, Size ou qualquer outro método do FileInfo.`,
	Run: runDiscardedStatError,
}

// runDiscardedStatError procura o padrão de Stat com erro descartado e uso
// posterior do FileInfo.
func runDiscardedStatError(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		parents := parentsOf(file)
		ast.Inspect(file, func(n ast.Node) bool {
			assign, ok := n.(*ast.AssignStmt)
			if !ok || len(assign.Lhs) != 2 || len(assign.Rhs) != 1 {
				return true
			}
			call, ok := ast.Unparen(assign.Rhs[0]).(*ast.CallExpr)
			if !ok || osStreamName(pass, call) == "" {
				return true
			}
			blank, ok := ast.Unparen(assign.Lhs[1]).(*ast.Ident)
			if !ok || blank.Name != "_" {
				return true
			}
			info, ok := ast.Unparen(assign.Lhs[0]).(*ast.Ident)
			if !ok || info.Name == "_" {
				return true
			}
			obj := pass.TypesInfo.ObjectOf(info)
			if obj == nil {
				return true
			}

			root := ast.Node(file)
			if fn := enclosingFunc(pass, assign.Pos()); fn != nil {
				root = fn.Body
			}
			for _, use := range usesOf(pass, obj, root) {
				sel, ok := parents[use].(*ast.SelectorExpr)
				if !ok || sel.X != use {
					continue
				}
				method := sel.Sel.Name
				if method != "Mode" && method != "Size" {
					continue
				}
				pass.Report(analysis.Diagnostic{
					Pos:      sel.Pos(),
					End:      sel.End(),
					Category: ruleDiscardedStatError,
					Message: fmt.Sprintf(
						"o erro de %s foi descartado: se Stat falhar, %s fica nil e %s.%s() causa panic; verifique o erro antes de usar o FileInfo",
						types.ExprString(call), info.Name, info.Name, method,
					),
				})
			}
			return true
		})
	}
	return nil, nil
}

// osStreamName devolve "Stdin", "Stdout" ou "Stderr" quando call é
// os.Stdin.Stat(), os.Stdout.Stat() ou os.Stderr.Stat(); caso contrário,
// devolve "".
func osStreamName(pass *analysis.Pass, call *ast.CallExpr) string {
	sel, ok := ast.Unparen(call.Fun).(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "Stat" {
		return ""
	}
	stream, ok := ast.Unparen(sel.X).(*ast.SelectorExpr)
	if !ok {
		return ""
	}
	v, ok := pass.TypesInfo.Uses[stream.Sel].(*types.Var)
	if !ok || v.Pkg() == nil || v.Pkg().Path() != "os" {
		return ""
	}
	switch v.Name() {
	case "Stdin", "Stdout", "Stderr":
		return v.Name()
	default:
		return ""
	}
}
