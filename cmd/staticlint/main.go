package main

import (
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/structtag"

	"honnef.co/go/tools/staticcheck"

	"github.com/kisielk/errcheck/errcheck"          // проверки наличия непроверенных ошибок
	"github.com/timakin/bodyclose/passes/bodyclose" // static analysis tool which checks whether res.Body is correctly closed.

	"github.com/eugene982/url-shortener/cmd/staticlint/internal/osexit" // свой анализатор.
)

// выбираем тут
// https://github.com/golangci/awesome-go-linters

func main() {

	mychecks := []*analysis.Analyzer{
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
		shift.Analyzer,
		bodyclose.Analyzer,
		errcheck.Analyzer,
		osexit.Analyzer,
	}

	// Добавляем по одному правилу из остальных групп
	checks := map[string]bool{
		"S1002": true, // Omit comparison with boolean constant

		"ST1005": true, // Incorrectly formatted error string
		"ST1008": true, // A function’s error value should be its last return value
		"ST1019": true, // Importing the same package multiple times
		"ST1020": true, // The documentation of an exported function should start with the function’s name

		"QF1010": true, // Convert slice of bytes to string when printing it
	}

	// добавляем анализаторы из staticcheck, SA
	for _, v := range staticcheck.Analyzers {
		if strings.HasPrefix(v.Analyzer.Name, "SA") {
			mychecks = append(mychecks, v.Analyzer)
		} else if checks[v.Analyzer.Name] {
			mychecks = append(mychecks, v.Analyzer)
		}
	}
	multichecker.Main(
		mychecks...,
	)
}
