package lint

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// ByteSlice detecta fatiamento de string por byte que chega à renderização.
var ByteSlice = &analysis.Analyzer{
	Name: "byteslice",
	Doc: `Detecta expressões de slice em string com limite não constante cujo resultado é impresso ou renderizado.

Fatiar uma string por byte (s[:n], s[n:]) pode cortar um rune ao meio e produzir UTF-8 inválido, que aparece como caractere quebrado no terminal. Use ansi.Truncate para cortar por largura de exibição.`,
	Run: runByteSlice,
}

// runByteSlice procura slices de string com limite variável que fluem para uma
// saída visível.
func runByteSlice(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		ctx := &flowCtx{pass: pass, file: file, parents: parentsOf(file)}
		ast.Inspect(file, func(n ast.Node) bool {
			slice, ok := n.(*ast.SliceExpr)
			if !ok {
				return true
			}
			if !isStringType(pass.TypesInfo.TypeOf(slice.X)) {
				return true
			}
			if isConstantInt(pass, slice.Low) && isConstantInt(pass, slice.High) && isConstantInt(pass, slice.Max) {
				return true
			}
			if !ctx.reachesOutput(slice, make(map[types.Object]bool), 0) {
				return true
			}

			pass.Report(analysis.Diagnostic{
				Pos:      slice.Pos(),
				End:      slice.End(),
				Category: ruleByteSlice,
				Message: fmt.Sprintf(
					"o slice %s corta a string por byte: se o limite cair no meio de um rune, o resultado vira UTF-8 inválido; use ansi.Truncate para cortar por largura de exibição",
					types.ExprString(slice),
				),
			})
			return true
		})
	}
	return nil, nil
}

// flowCtx rastreia o caminho de uma expressão até uma saída visível.
type flowCtx struct {
	pass    *analysis.Pass
	file    *ast.File
	parents map[ast.Node]ast.Node
}

// reachesOutput informa se expr (ou o valor que ela alimenta) termina em uma
// chamada de impressão/renderização, em uma concatenação de strings ou no
// retorno de uma função de renderização.
//
// O rastreamento sobe pela árvore e, quando encontra uma atribuição, segue os
// usos da variável. O limite de profundidade evita ciclos.
func (c *flowCtx) reachesOutput(expr ast.Expr, seen map[types.Object]bool, depth int) bool {
	if depth > 8 {
		return false
	}
	switch parent := c.parents[expr].(type) {
	case *ast.ParenExpr:
		return c.reachesOutput(parent, seen, depth+1)
	case *ast.BinaryExpr:
		// Concatenação de strings: o resultado é texto que será exibido.
		return parent.Op == token.ADD &&
			(isStringType(c.pass.TypesInfo.TypeOf(parent.X)) || isStringType(c.pass.TypesInfo.TypeOf(parent.Y)))
	case *ast.CallExpr:
		if isPrintOrRenderCall(c.pass, parent) {
			return true
		}
		// A chamada devolve string e o resultado segue para a saída
		// (por exemplo, strings.ToUpper(s[:n]) dentro de fmt.Print).
		if isStringType(c.pass.TypesInfo.TypeOf(parent)) {
			return c.reachesOutput(parent, seen, depth+1)
		}
		return false
	case *ast.AssignStmt:
		return c.objectReachesOutput(assignedObject(c.pass, parent, expr), seen, depth+1)
	case *ast.ValueSpec:
		return c.objectReachesOutput(assignedSpecObject(c.pass, parent, expr), seen, depth+1)
	case *ast.ReturnStmt:
		fn := enclosingFunc(c.pass, parent.Pos())
		return fn != nil && renderFuncName(fn.Name)
	}
	return false
}

// objectReachesOutput segue os usos de obj dentro da função que o declara.
func (c *flowCtx) objectReachesOutput(obj types.Object, seen map[types.Object]bool, depth int) bool {
	if obj == nil || seen[obj] {
		return false
	}
	seen[obj] = true

	root := ast.Node(c.file)
	if fn := enclosingFunc(c.pass, obj.Pos()); fn != nil {
		root = fn.Body
	}
	for _, use := range usesOf(c.pass, obj, root) {
		if c.reachesOutput(use, seen, depth+1) {
			return true
		}
	}
	return false
}

// assignedObject devolve a variável que recebe expr em uma atribuição.
func assignedObject(pass *analysis.Pass, assign *ast.AssignStmt, expr ast.Expr) types.Object {
	if len(assign.Rhs) != len(assign.Lhs) {
		return nil
	}
	for i, rhs := range assign.Rhs {
		if ast.Unparen(rhs) != ast.Unparen(expr) {
			continue
		}
		id, ok := assign.Lhs[i].(*ast.Ident)
		if !ok || id.Name == "_" {
			return nil
		}
		return pass.TypesInfo.ObjectOf(id)
	}
	return nil
}

// assignedSpecObject devolve a variável que recebe expr em um var/const.
func assignedSpecObject(pass *analysis.Pass, spec *ast.ValueSpec, expr ast.Expr) types.Object {
	if len(spec.Values) != len(spec.Names) {
		return nil
	}
	for i, value := range spec.Values {
		if ast.Unparen(value) != ast.Unparen(expr) {
			continue
		}
		return pass.TypesInfo.ObjectOf(spec.Names[i])
	}
	return nil
}
