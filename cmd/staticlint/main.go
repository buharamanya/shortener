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

// Config конфигурация анализаторов
type Config struct {
	StaticCheckAnalyzers map[string]bool
	CustomAnalyzers      []*analysis.Analyzer
}

// exitCheckAnalyzer анализатор вызова os.Exit() функции в main пакете.
var exitCheckAnalyzer = &analysis.Analyzer{
	Name: "exitcheck",                      // имя анализатора.
	Doc:  "check os.Exit in main function", // текст с описанием работы анализатора.
	Run:  run,                              // функция, которая отвечает за анализ исходного кода.
}

func main() {
	config := Config{
		StaticCheckAnalyzers: map[string]bool{
			"SA":     true,
			"S1006":  true,
			"ST1012": true,
		},
		CustomAnalyzers: []*analysis.Analyzer{
			copylock.Analyzer,  // checks for locks erroneously passed by value.
			defers.Analyzer,    // checks for common mistakes in defer statements.
			printf.Analyzer,    // checks consistency of Printf format strings and arguments.
			shadow.Analyzer,    // checks for shadowed variables.
			structtag.Analyzer, // checks struct field tags are well formed.
			unmarshal.Analyzer, // checks for passing non-pointer or non-interface types to unmarshal and decode functions.
			errcheck.Analyzer,  // checks unchecked errors in Go code.
			gocc.Analyzer,      // checks cyclomatic complexity of go functions.
			exitCheckAnalyzer,  // checks call os.Exit() function in main package.
		},
	}

	mychecks := config.CustomAnalyzers

	for _, v := range staticcheck.Analyzers {
		if config.StaticCheckAnalyzers[v.Analyzer.Name] {
			mychecks = append(mychecks, v.Analyzer)
		}
	}

	multichecker.Main(mychecks...)
}

// run функция, которая отвечает за анализ исходного кода.
func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if !isMainPackage(file) {
			continue
		}

		if !hasOSImport(file.Imports) {
			continue
		}

		checkForOSExitCalls(pass, file)
	}
	return nil, nil
}

// isMainPackage проверяет, является ли файл частью main пакета
func isMainPackage(file *ast.File) bool {
	return file.Name.Name == "main"
}

// hasOSImport проверяет среди импортов наличие пакета os
func hasOSImport(imports []*ast.ImportSpec) bool {
	for _, importSpec := range imports {
		if importSpec.Path != nil && importSpec.Path.Value == "\"os\"" {
			return true
		}
	}
	return false
}

// checkForOSExitCalls проверяет наличие вызовов os.Exit в AST дереве файла
func checkForOSExitCalls(pass *analysis.Pass, file *ast.File) {
	ast.Inspect(file, func(node ast.Node) bool {
		callExpr, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}

		if isOSExitCall(callExpr) {
			reportOSExitCall(pass, callExpr)
		}

		return true
	})
}

// isOSExitCall проверяет, является ли вызов выражением os.Exit
func isOSExitCall(callExpr *ast.CallExpr) bool {
	selector, ok := callExpr.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	ident, ok := selector.X.(*ast.Ident)
	if !ok {
		return false
	}

	return ident.Name == "os" && selector.Sel.Name == "Exit"
}

// reportOSExitCall сообщает о найденном вызове os.Exit
func reportOSExitCall(pass *analysis.Pass, callExpr *ast.CallExpr) {
	if selector, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
		if ident, ok := selector.X.(*ast.Ident); ok {
			pass.Reportf(ident.NamePos, "call os.Exit function")
		}
	}
}
