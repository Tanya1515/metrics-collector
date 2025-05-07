// Multichecker is used to analyze code and find missprints, some simple errors. 
// Multichecker consists of several connected analyzers: 
// 1. appends - Analyzer that detects if there is only one variable in append.
// 2. errorsas - Analyzer that checks that the second argument to errors.As is a pointer to a type implementing error.
// 3. httpresponse - Analyzer that checks for mistakes using HTTP responses.
// 4. nilfunc - Analyzer that checks for useless comparisons against nil.
// 5. printf - Analyzer that checks consistency of Printf format strings and arguments.
// 6. shadow - Analyzer that checks for shadowed variables. 
// 7. unmarshal - Analyzer that checks for passing non-pointer or non-interface types to unmarshal and decode functions.
// 8. unreachable - Analyzer that checks for unreachable code.
// 9. staticcheck  - open source static analyzator with different types of checks.  
// 10. staticcheck ST1000 - Analyzer, that checks for unusing variables in code.
// 11. staticcheck QF1003 - Analyzer, that checks for unusing imports. 
// 12. ExitCheckAnalyzer - my static analyzator for checking if main contains os.Exit call.
// 13. staticcheck SA - Analyzer, that checks with tests, unusefull or unreachable code, efficiency problems and etc.
// 14. ineffassign - Analyzer, that checks for ineffective assignment.
// 15. nilerr - Analyzer, that checks cases when function returns nil as an error and nil as another varibles. 

// For running multichecker do the following steps: 
// a) go from the project root into directory staticlint.
// b) run 'go build' for building the multichecker. As a result binary file with name staticlint will be in the directory.
// c) run the binary file with path to processing directory or file: ./staticlint ../server.

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

	"github.com/gordonklaus/ineffassign/pkg/ineffassign"
	"github.com/gostaticanalysis/nilerr"
)

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
		nilerr.Analyzer,
		ineffassign.Analyzer,
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
