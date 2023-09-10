// Package osexit - анализатор, запрещающий использовать
// прямой вызов os.Exit в функции main пакета main.
package osexit

import (
	"go/ast"
	"regexp"

	"golang.org/x/tools/go/analysis"
)

// Analyzer информация о б анализаторе
var Analyzer = &analysis.Analyzer{
	Name: "osexit",                                               // Имя анализатора
	Doc:  "check call os.Exit() in main function (package main)", // Текст описания работы анализатора
	Run:  run,                                                    // Функция, которая отвечает за анадиз исходного кода
}

// Исключаем сгенерированные файла
var excludeGenFileComent = regexp.MustCompile(`^// Code generated .* DO NOT EDIT\.$`)

func run(pass *analysis.Pass) (any, error) {

	for _, file := range pass.Files {

		var (
			isMainPkg bool
			isMainFn  bool
		)

		// 	функцией ast.Inspect проходим по всем узлам AST
		ast.Inspect(file, func(node ast.Node) bool {

			if excludeFile(file) {
				return false
			}

			switch x := node.(type) {
			case *ast.File: // проверка имени пакета
				isMainPkg = x.Name.Name == "main"

			case *ast.FuncDecl: // проверка имени функции
				isMainFn = x.Name.Name == "main"

			case *ast.ExprStmt: // проверка выражения на вызов нужной функции
				if isMainFn && isMainPkg && isOsExitFunc(x) {
					pass.Reportf(x.Pos(), "call os.Exit() in main function")
				}
			}
			return true
		})
	}

	return nil, nil
}

func isOsExitFunc(expr *ast.ExprStmt) bool {
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

func excludeFile(file *ast.File) bool {
	for _, c := range file.Comments {
		for _, comment := range c.List {
			if excludeGenFileComent.MatchString(comment.Text) {
				return true
			}
		}
	}
	return false
}
