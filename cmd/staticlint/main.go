// подключить два публичных анализатора и описать документацию
// проверить, что мой код проходит все проверки подключенных анализаторов
package main

import (
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/appends"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"

	"honnef.co/go/tools/staticcheck"
)

type Analyzer struct {
	Analyzer *analysis.Analyzer
}

func main() {
	mychecks := []*analysis.Analyzer{
		ExitCheckAnalyzer,
		appends.Analyzer,
		errorsas.Analyzer,
		httpresponse.Analyzer,
		nilfunc.Analyzer,
		structtag.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		printf.Analyzer,
		shadow.Analyzer,
	}

	checks := map[string]bool{
		"S1040":  true, // Type assertion to current type
		"ST1000": true, // Incorrect or missing package comment
		"QF1003": true, // Convert if/else-if chain to tagged switch
	}

	for _, v := range staticcheck.Analyzers {
		if strings.Contains(v.Analyzer.Name, "SA") || checks[v.Analyzer.Name] {
			mychecks = append(mychecks, v.Analyzer)
		}

	}

	multichecker.Main(
		mychecks...,
	)
}
