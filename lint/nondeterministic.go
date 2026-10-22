package lint

import (
	"fmt"
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
)

// NondeterministicRender detecta iteração sobre map dentro de funções de
// renderização.
var NondeterministicRender = &analysis.Analyzer{
	Name: "nondeterministicrender",
	Doc: `Detecta "for ... := range map" dentro de funções Render*, View* e String.

A ordem de iteração de um map não é especificada e muda entre execuções, então a saída renderizada fica diferente a cada vez. Colete as chaves, ordene com sort.Strings/slices.Sort e itere sobre a lista ordenada.`,
	Run: runNondeterministicRender,
}

// runNondeterministicRender procura ranges sobre map em funções cujo nome
// indica renderização.
func runNondeterministicRender(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil || !renderFuncName(fn.Name.Name) {
				continue
			}
			if fn.Name.Name == "String" && !isStringMethod(fn.Type) {
				continue
			}

			ast.Inspect(fn.Body, func(n ast.Node) bool {
				loop, ok := n.(*ast.RangeStmt)
				if !ok || loop.Body == nil {
					return true
				}
				if !isMapType(pass.TypesInfo.TypeOf(loop.X)) {
					return true
				}
				// Um loop que apenas coleta chaves para ordenar depois não
				// produz saída e não deve ser sinalizado.
				if !hasOutputEvidence(pass, loop.Body) {
					return true
				}
				pass.Report(analysis.Diagnostic{
					Pos:      loop.Pos(),
					End:      loop.End(),
					Category: ruleNondeterministic,
					Message: fmt.Sprintf(
						"iteração sobre map em %s produz saída em ordem aleatória, porque a ordem de range em map não é especificada; colete as chaves, ordene com sort.Strings ou slices.Sort e itere sobre a lista ordenada",
						fn.Name.Name,
					),
				})
				return true
			})
		}
	}
	return nil, nil
}

// isStringMethod informa se o método tem a assinatura de fmt.Stringer:
// nenhum parâmetro e um único resultado string.
func isStringMethod(typ *ast.FuncType) bool {
	if typ.Params != nil && len(typ.Params.List) > 0 {
		return false
	}
	if typ.Results == nil || len(typ.Results.List) != 1 {
		return false
	}
	return isStringTypeExpr(typ.Results.List[0].Type)
}

// isStringTypeExpr reconhece a forma sintática do tipo string. A checagem é
// suficiente aqui porque String() de fmt.Stringer precisa devolver string.
func isStringTypeExpr(expr ast.Expr) bool {
	id, ok := expr.(*ast.Ident)
	return ok && id.Name == "string"
}

// hasOutputEvidence informa se o corpo do loop produz saída visível: chamadas
// de impressão/escrita/formatação, concatenação de strings ou += de string.
//
// O critério separa o loop que renderiza do loop que apenas coleta chaves
// para ordenar em seguida.
func hasOutputEvidence(pass *analysis.Pass, body *ast.BlockStmt) bool {
	found := false
	ast.Inspect(body, func(n ast.Node) bool {
		if found {
			return false
		}
		switch node := n.(type) {
		case *ast.CallExpr:
			if isOutputCall(node) {
				found = true
				return false
			}
		case *ast.BinaryExpr:
			if node.Op == token.ADD &&
				(isStringType(pass.TypesInfo.TypeOf(node.X)) || isStringType(pass.TypesInfo.TypeOf(node.Y))) {
				found = true
				return false
			}
		case *ast.AssignStmt:
			if node.Tok == token.ADD_ASSIGN && len(node.Lhs) == 1 && isStringType(pass.TypesInfo.TypeOf(node.Lhs[0])) {
				found = true
				return false
			}
		}
		return true
	})
	return found
}
