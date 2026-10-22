package lint

import (
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// Identificadores estáveis das regras. Eles aparecem em
// analysis.Diagnostic.Category, no README e nos testes.
const (
	ruleByteLengthPadding  = "bytelength-padding"
	ruleByteSlice          = "byteslice"
	ruleRangeIndexArith    = "range-index-arith"
	ruleNondeterministic   = "nondeterministic-render"
	ruleSilentExec         = "silent-exec"
	ruleDiscardedStatError = "discarded-stat-error"
)

// isPkgFunc informa se call invoca a função name do pacote path.
//
// A resolução usa a informação de tipos, então imports com alias continuam
// sendo reconhecidos e uma função local homônima não gera falso positivo.
func isPkgFunc(pass *analysis.Pass, call *ast.CallExpr, path, name string) bool {
	fn := calledFunc(pass, call)
	if fn == nil {
		return false
	}
	return fn.Name() == name && fn.Pkg() != nil && fn.Pkg().Path() == path
}

// calledFunc devolve a função invocada por call, ou nil quando o alvo não é
// uma função (por exemplo, uma conversão de tipo ou um builtin).
func calledFunc(pass *analysis.Pass, call *ast.CallExpr) *types.Func {
	var id *ast.Ident
	switch fun := ast.Unparen(call.Fun).(type) {
	case *ast.Ident:
		id = fun
	case *ast.SelectorExpr:
		id = fun.Sel
	default:
		return nil
	}
	fn, _ := pass.TypesInfo.Uses[id].(*types.Func)
	return fn
}

// isStringType informa se t é uma string ou um tipo nomeado cujo tipo
// subjacente é string.
func isStringType(t types.Type) bool {
	if t == nil {
		return false
	}
	basic, ok := t.Underlying().(*types.Basic)
	return ok && basic.Kind() == types.String
}

// isMapType informa se t é um map ou um tipo nomeado cujo tipo subjacente é
// map.
func isMapType(t types.Type) bool {
	if t == nil {
		return false
	}
	_, ok := t.Underlying().(*types.Map)
	return ok
}

// parentsOf monta o mapa de pai para cada nó da árvore sintática.
func parentsOf(root ast.Node) map[ast.Node]ast.Node {
	parents := make(map[ast.Node]ast.Node)
	var stack []ast.Node
	ast.Inspect(root, func(n ast.Node) bool {
		if n == nil {
			stack = stack[:len(stack)-1]
			return false
		}
		if len(stack) > 0 {
			parents[n] = stack[len(stack)-1]
		}
		stack = append(stack, n)
		return true
	})
	return parents
}

// funcInfo descreve a função que contém uma posição.
type funcInfo struct {
	// Name é o nome da função; vazio para funções literais.
	Name string
	// Body é o corpo da função; nunca é nil.
	Body *ast.BlockStmt
	// Start e End delimitam a declaração inteira.
	Start, End token.Pos
}

// enclosingFunc devolve a função mais interna que contém pos, ou nil quando
// pos está fora de qualquer corpo de função.
func enclosingFunc(pass *analysis.Pass, pos token.Pos) *funcInfo {
	var best *funcInfo
	consider := func(name string, node ast.Node, body *ast.BlockStmt) {
		if body == nil || pos < node.Pos() || pos > node.End() {
			return
		}
		if best == nil || node.End()-node.Pos() < best.End-best.Start {
			best = &funcInfo{Name: name, Body: body, Start: node.Pos(), End: node.End()}
		}
	}
	for _, file := range pass.Files {
		if pos < file.Pos() || pos > file.End() {
			continue
		}
		ast.Inspect(file, func(n ast.Node) bool {
			switch fn := n.(type) {
			case *ast.FuncDecl:
				consider(fn.Name.Name, fn, fn.Body)
			case *ast.FuncLit:
				consider("", fn, fn.Body)
			}
			return true
		})
	}
	return best
}

// usesOf devolve todos os identificadores que se referem a obj dentro de root.
func usesOf(pass *analysis.Pass, obj types.Object, root ast.Node) []*ast.Ident {
	var uses []*ast.Ident
	ast.Inspect(root, func(n ast.Node) bool {
		id, ok := n.(*ast.Ident)
		if ok && pass.TypesInfo.Uses[id] == obj {
			uses = append(uses, id)
		}
		return true
	})
	return uses
}

// renderFuncName informa se name identifica uma função de renderização: as
// que começam com Render ou View e o método String.
func renderFuncName(name string) bool {
	return strings.HasPrefix(name, "Render") ||
		strings.HasPrefix(name, "View") ||
		name == "String"
}

// isPrintOrRenderCall informa se call produz saída visível: funções da
// família Print/Sprintf (inclusive métodos como lipgloss Render).
func isPrintOrRenderCall(pass *analysis.Pass, call *ast.CallExpr) bool {
	name := calleeName(call)
	lower := strings.ToLower(name)
	return strings.Contains(lower, "print") || strings.Contains(lower, "render")
}

// isOutputCall informa se call escreve ou formata saída, incluindo Write e
// Format, usados para montar buffers de renderização.
func isOutputCall(call *ast.CallExpr) bool {
	lower := strings.ToLower(calleeName(call))
	for _, marker := range []string{"print", "render", "write", "format"} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

// calleeName devolve o nome da função invocada por call, sem qualificação.
func calleeName(call *ast.CallExpr) string {
	switch fun := ast.Unparen(call.Fun).(type) {
	case *ast.Ident:
		return fun.Name
	case *ast.SelectorExpr:
		return fun.Sel.Name
	default:
		return ""
	}
}

// isConstantInt informa se expr é nil ou uma expressão constante de valor
// inteiro conhecido em tempo de compilação.
func isConstantInt(pass *analysis.Pass, expr ast.Expr) bool {
	if expr == nil {
		return true
	}
	return pass.TypesInfo.Types[expr].Value != nil
}
