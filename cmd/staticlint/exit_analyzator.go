package main

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

type Analyzer struct {
	Analyzer *analysis.Analyzer
}

var ExitCheckAnalyzer = &analysis.Analyzer{
	Name: "exitAnalyzer",
	Doc:  "check for calling os.Exit in main",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {

	for _, file := range pass.Files {
		fileInfo := pass.Fset.File(file.Pos())
		pass.Reportf(file.Pos(), fileInfo.Name())
		if strings.Contains(fileInfo.Name(), "/main.go") {
			ast.Inspect(file, func(node ast.Node) bool {
				if funcDecl, ok := node.(*ast.FuncDecl); ok {
					if funcDecl.Name.Name != "main" {
						return false
					}
				}
				callExpr, ok := node.(*ast.CallExpr)
				if ok {
					selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
					if ok {
						ident, ok := selExpr.X.(*ast.Ident)
						if ok && (ident.Name == "os") && (selExpr.Sel.Name == "Exit") {
							pass.Reportf(ident.Pos(), "os.Exit() can not be used in main function of main.go")
						}
					}
				}
				return true
			})
		}
	}
	return nil, nil
}
