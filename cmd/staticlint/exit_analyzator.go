// проверить, что анализатор работает - какой-нибудь тупой пример со сторонней функцией
package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var ExitCheckAnalyzer = &analysis.Analyzer{
	Name: "exitAnalyzer",
	Doc:  "check for calling os.Exit in main",
	Run:  runExit,
}

// добавить проверку того, что os.Exit вызывается внутри функции main

func runExit(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if file.Name.Name != "main.go" {
			continue
		}
		ast.Inspect(file, func(node ast.Node) bool {
			callExpr, ok := node.(*ast.CallExpr)
			if ok {
				selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
				if ok {
					ident, ok := selExpr.X.(*ast.Ident)
					if ok && (ident.Name == "os") && (selExpr.Sel.Name == "Exit") {
						pass.Reportf(ident.Pos(), "os.Exit() can be used in main function of main.go")
					}
				}
			}
			return true
		})
	}
	return nil, nil
}
