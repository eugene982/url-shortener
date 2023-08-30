// Package osexit - анализатор, запрещающий использовать
// прямой вызов os.Exit в функции main пакета main.
package osexit

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
)

// Analyzer информация о б анализаторе
var Analyzer = &analysis.Analyzer{
	Name: "osexit",                                             // Имя анализатора
	Doc:  "check call os.Exit in main function (package main)", // Текст описания работы анализатора
	Run:  run,                                                  // Функция, которая отвечает за анадиз исходного кода
}

func run(pass *analysis.Pass) (any, error) {

	for _, file := range pass.Files {

		var (
			isMainPkg bool
			isMainFn  bool
			ignorePos token.Pos
		)

		// 	функцией ast.Inspect проходим по всем узлам AST
		ast.Inspect(file, func(node ast.Node) bool {

			switch x := node.(type) {
			case *ast.File: // проверка имени пакета
				isMainPkg = x.Name.Name == "main"

			case *ast.FuncDecl: // проверка имени функции
				isMainFn = x.Name.Name == "main"

			case *ast.ExprStmt: // проверка выражения на вызов нужной функции
				if isMainFn && isMainPkg && isOsExitFunc(x, ignorePos) {
					pass.Reportf(x.Pos(), "os.Exit in main func")
				}
			}
			return true
		})
	}

	return nil, nil
}

func isOsExitFunc(expr *ast.ExprStmt, ignore token.Pos) bool {
	call, ok := expr.X.(*ast.CallExpr)
	if !ok {
		return false
	}
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	if x, ok := selector.X.(*ast.Ident); !ok || x.Name != "os" {
		return false
	}

	return selector.Sel.Name == "Exit"
}
