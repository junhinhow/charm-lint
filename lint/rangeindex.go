package lint

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// RangeIndexArith detecta aritmética com o índice de um range sobre string.
var RangeIndexArith = &analysis.Analyzer{
	Name: "rangeindexarith",
	Doc: `Detecta aritmética com o índice de "for i := range s" quando s é string.

Em uma string o índice do range é o offset em bytes, não o índice do rune. Usar i+1, i-1 ou comparações com i pode cair no meio de um rune e produzir resultados errados. Itere com "for _, r := range s" ou converta a string para []rune.`,
	Run: runRangeIndexArith,
}

// runRangeIndexArith procura usos aritméticos do índice de um range sobre
// string.
func runRangeIndexArith(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		parents := parentsOf(file)
		ast.Inspect(file, func(n ast.Node) bool {
			loop, ok := n.(*ast.RangeStmt)
			if !ok || loop.Key == nil || loop.Value != nil || loop.Body == nil {
				return true
			}
			if !isStringType(pass.TypesInfo.TypeOf(loop.X)) {
				return true
			}
			key, ok := loop.Key.(*ast.Ident)
			if !ok || key.Name == "_" {
				return true
			}
			obj := pass.TypesInfo.ObjectOf(key)
			if obj == nil {
				return true
			}

			for _, use := range usesOf(pass, obj, loop.Body) {
				bin, ok := parentBinaryExpr(parents, use)
				if !ok || (bin.X != use && bin.Y != use) {
					continue
				}
				switch bin.Op {
				case token.ADD, token.SUB, token.LSS, token.GTR, token.LEQ, token.GEQ, token.EQL, token.NEQ:
				default:
					continue
				}
				pass.Report(analysis.Diagnostic{
					Pos:      bin.Pos(),
					End:      bin.End(),
					Category: ruleRangeIndexArith,
					Message: fmt.Sprintf(
						"%s usa o índice de range sobre string, que é offset de byte e não índice de rune: a conta pode cair no meio de um rune; itere com \"for _, r := range %s\" ou converta para []rune",
						types.ExprString(bin), types.ExprString(loop.X),
					),
				})
			}
			return true
		})
	}
	return nil, nil
}

// parentBinaryExpr sobe pelos parênteses até a expressão binária que contém
// use.
func parentBinaryExpr(parents map[ast.Node]ast.Node, use ast.Node) (*ast.BinaryExpr, bool) {
	parent := parents[use]
	for {
		if paren, ok := parent.(*ast.ParenExpr); ok {
			parent = parents[paren]
			continue
		}
		bin, ok := parent.(*ast.BinaryExpr)
		return bin, ok
	}
}
