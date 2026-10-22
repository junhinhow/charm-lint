package lint

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// ByteLengthPadding detecta padding calculado com len ou com
// utf8.RuneCountInString.
var ByteLengthPadding = &analysis.Analyzer{
	Name: "bytelengthpadding",
	Doc: `Detecta strings.Repeat usado para padding com contagem em bytes (len) ou em runes (utf8.RuneCountInString).

len conta bytes: "ação" ocupa 6 bytes e 4 células, então o padding sai maior do que o esperado. utf8.RuneCountInString conta runes, mas CJK e emoji ocupam duas células, então o padding sai menor. Nos dois casos use ansi.StringWidth para medir a largura de exibição.`,
	Run: runByteLengthPadding,
}

// measureKind identifica como uma expressão mede o tamanho de uma string.
type measureKind int

const (
	measureNone measureKind = iota
	measureBytes
	measureRunes
)

// runByteLengthPadding procura strings.Repeat cujo segundo argumento mede uma
// string por bytes ou por runes.
func runByteLengthPadding(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok || len(call.Args) != 2 || !isPkgFunc(pass, call, "strings", "Repeat") {
				return true
			}
			kind, arg := lengthMeasurement(pass, call.Args[1])
			if kind == measureNone {
				return true
			}

			var msg string
			switch kind {
			case measureBytes:
				msg = fmt.Sprintf("padding calculado com len(%s): len conta bytes, então texto com acento ou emoji ocupa mais bytes do que células e o alinhamento quebra; meça a largura de exibição com ansi.StringWidth", arg)
			case measureRunes:
				msg = fmt.Sprintf("padding calculado com utf8.RuneCountInString(%s): contagem de runes não é largura de exibição (CJK e emoji ocupam duas células); use ansi.StringWidth", arg)
			}
			pass.Report(analysis.Diagnostic{
				Pos:      call.Pos(),
				End:      call.End(),
				Category: ruleByteLengthPadding,
				Message:  msg,
			})
			return true
		})
	}
	return nil, nil
}

// lengthMeasurement identifica se expr mede uma string por bytes ou por runes.
//
// Aceita a medição direta e a forma clássica de padding "largura - medição",
// que é como o preenchimento costuma ser calculado.
func lengthMeasurement(pass *analysis.Pass, expr ast.Expr) (measureKind, string) {
	expr = ast.Unparen(expr)
	if kind, arg, ok := directMeasurement(pass, expr); ok {
		return kind, arg
	}
	if bin, ok := expr.(*ast.BinaryExpr); ok && bin.Op == token.SUB {
		if kind, arg, ok := directMeasurement(pass, ast.Unparen(bin.Y)); ok {
			return kind, arg
		}
	}
	return measureNone, ""
}

// directMeasurement reconhece len(s) e utf8.RuneCountInString(s) quando o
// argumento é uma string.
func directMeasurement(pass *analysis.Pass, expr ast.Expr) (measureKind, string, bool) {
	call, ok := expr.(*ast.CallExpr)
	if !ok || len(call.Args) != 1 {
		return measureNone, "", false
	}
	if id, ok := ast.Unparen(call.Fun).(*ast.Ident); ok && id.Name == "len" {
		if _, isBuiltin := pass.TypesInfo.Uses[id].(*types.Builtin); isBuiltin && isStringType(pass.TypesInfo.TypeOf(call.Args[0])) {
			return measureBytes, types.ExprString(call.Args[0]), true
		}
	}
	if isPkgFunc(pass, call, "unicode/utf8", "RuneCountInString") {
		return measureRunes, types.ExprString(call.Args[0]), true
	}
	return measureNone, "", false
}
