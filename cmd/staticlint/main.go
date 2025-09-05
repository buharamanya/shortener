// staticlint инструмент для статического анализа Go кода.
// Перед использованием неообходимо собрать пакет командой go build . (в директории cmd/staticlint)
// Использование: staticlint [-flag] [package]
// Для вызова подсказки выполните 'staticlint help' или 'staticlint help name'
// для более детальнного описания анализатора.
package main

import (
	"go/ast"

	"github.com/kisielk/errcheck/errcheck"
	"github.com/knsh14/gocc"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/defers"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"honnef.co/go/tools/staticcheck"
)

// exitCheckAnalyzer анализатор вызова os.Exit() функции в main пакете.
var exitCheckAnalyzer = &analysis.Analyzer{
	Name: "exitcheck",                      // имя анализатора.
	Doc:  "check os.Exit in main function", // текст с описанием работы анализатора.
	Run:  run,                              // функция, которая отвечает за анализ исходного кода.
}

func main() {
	staticchecks := map[string]bool{
		"SA":     true,
		"S1006":  true,
		"ST1012": true,
	}

	mychecks := []*analysis.Analyzer{
		copylock.Analyzer,  // checks for locks erroneously passed by value.
		defers.Analyzer,    // checks for common mistakes in defer statements.
		printf.Analyzer,    // checks consistency of Printf format strings and arguments.
		shadow.Analyzer,    // checks for shadowed variables.
		structtag.Analyzer, // checks struct field tags are well formed.
		unmarshal.Analyzer, // checks for passing non-pointer or non-interface types to unmarshal and decode functions.
		errcheck.Analyzer,  // checks unchecked errors in Go code.
		gocc.Analyzer,      // checks cyclomatic complexity of go functions.
		exitCheckAnalyzer,  // checks call os.Exit() function in main package.
	}

	for _, v := range staticcheck.Analyzers {
		if staticchecks[v.Analyzer.Name] {
			mychecks = append(mychecks, v.Analyzer)
		}
	}

	multichecker.Main(
		mychecks...,
	)
}

// run функция, которая отвечает за анализ исходного кода.
func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if file.Name.Name != "main" {
			continue
		}

		if osImportAbsent(file, file.Imports) {
			continue
		}

		ast.Inspect(file, func(node ast.Node) bool {
			if call, ok := node.(*ast.CallExpr); ok {
				if selector, ok := call.Fun.(*ast.SelectorExpr); ok {
					if p, ok := selector.X.(*ast.Ident); ok {
						if p.Name == "os" && selector.Sel.Name == "Exit" {
							pass.Reportf(p.NamePos, "call os.Exit function")
						}
					}
				}
			}

			return true
		})
	}
	return nil, nil
}

// osImportAbsent проверяет среди импортов наличие пакета os
func osImportAbsent(f *ast.File, imports []*ast.ImportSpec) bool {
	for _, importSpec := range imports {
		if _, ok := f.Scope.Objects["tests"]; ok {
			// Исключил попадаение
			// ~/Library/Caches/go-build/ed/ed7616fad552f6e19a3945b71c97ae2f5f9ee5dbf7d548ed82c1d080d99bb245-d:47:2
			continue
		}

		lit := importSpec.Path
		if lit != nil && lit.Value == "\"os\"" {
			return false
		}
	}

	return true
}
