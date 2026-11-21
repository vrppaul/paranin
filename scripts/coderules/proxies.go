package main

import (
	"go/ast"
	"go/types"
	"slices"
	"strings"

	"golang.org/x/tools/go/analysis"
)

const noProxiesDescription = `refuse a function that only passes its call on

A function whose body is one call, or one call held and returned, where what is called and each thing handed
to it is a name, a field or a literal as it stands, does no work of its own: it hides from its reader what
happens. It is called in place, or its reason stands above it as the directive //wrapper:<why>, which no doc
prints. Taken from stint's scripts/check_proxies.py, and fitted to Go: a conversion names a type and is no
call; a test, and a method of a type that implements an interface asking for it, are found by their names;
an exported function that reads an unexported field is its package's way to that field, a getter, so no
caller outside could call it in place.`

const proxyMessage = "%s only passes its call on: call it in place, or say why on the line above, //wrapper:<why>"

// The directive that gives a function's reason to pass its call on
const wrapperDirective = "//wrapper:"

// The words a test function's name opens with, by which the go command finds it
func testFunctionPrefixes() []string {
	return []string{"Test", "Benchmark", "Fuzz", "Example"}
}

func noProxiesAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{Name: "noproxies", Doc: noProxiesDescription, Run: runNoProxies}
}

func runNoProxies(pass *analysis.Pass) (any, error) {
	interfaces := interfacesInScope(pass.Pkg)
	for _, file := range writtenFiles(pass) {
		for _, declaration := range file.Decls {
			function, isFunction := declaration.(*ast.FuncDecl)
			if isFunction && isProxy(pass, file, function, interfaces) {
				pass.Reportf(function.Name.Pos(), proxyMessage, function.Name.Name)
			}
		}
	}
	return nil, nil
}

// isProxy tells whether a function only passes its call on and nothing excuses it: no reason above it, no name it is
// found by, no unexported field it is the exported way to.
func isProxy(pass *analysis.Pass, file *ast.File, function *ast.FuncDecl, interfaces []*types.Interface) bool {
	if function.Body == nil || hasReason(function) || isFoundByName(pass, file, function, interfaces) {
		return false
	}
	call := passedOnCall(pass, function.Body.List)
	return call != nil && (!isExportedWay(pass, function) || !readsUnexportedField(pass, call))
}

// isExportedWay tells whether a caller outside the package reaches the function: it is exported, and a method's
// type is too.
func isExportedWay(pass *analysis.Pass, function *ast.FuncDecl) bool {
	method, isFunc := pass.TypesInfo.Defs[function.Name].(*types.Func)
	if !isFunc || !function.Name.IsExported() {
		return false
	}
	receiver := method.Signature().Recv()
	if receiver == nil {
		return true
	}
	named, isNamed := types.Unalias(dereferenced(receiver.Type())).(*types.Named)
	return isNamed && named.Obj().Exported()
}

// dereferenced is the type a pointer points to, or the type itself.
func dereferenced(held types.Type) types.Type {
	if pointer, isPointer := held.(*types.Pointer); isPointer {
		return pointer.Elem()
	}
	return held
}

// hasReason tells whether the function's doc holds the directive that gives its reason; gofmt keeps a directive
// only with words after it.
func hasReason(function *ast.FuncDecl) bool {
	if function.Doc == nil {
		return false
	}
	return slices.ContainsFunc(function.Doc.List, func(comment *ast.Comment) bool {
		return strings.HasPrefix(comment.Text, wrapperDirective) && len(comment.Text) > len(wrapperDirective)
	})
}

// isFoundByName tells whether what calls the function finds it by its name: a test the go command runs, or a method
// of a type that implements an interface asking for it.
func isFoundByName(pass *analysis.Pass, file *ast.File, function *ast.FuncDecl, interfaces []*types.Interface) bool {
	if function.Recv == nil {
		return isTestFile(pass, file) && startsAsTest(function.Name.Name)
	}
	return isAskedFor(pass.TypesInfo.Defs[function.Name], interfaces)
}

func startsAsTest(name string) bool {
	for _, prefix := range testFunctionPrefixes() {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}

// passedOnCall is the call a body only passes on: one call, alone or held and returned, on what stands unchanged;
// nil where the body does work of its own.
func passedOnCall(pass *analysis.Pass, statements []ast.Stmt) *ast.CallExpr {
	switch len(statements) {
	case callAlone:
		return unchangedCall(pass, onlyExpression(statements[0]))
	case heldThenReturned:
		return heldAndReturned(pass, statements[0], statements[1])
	}
	return nil
}

// The lengths of a body that passes its call on: the call alone, or the call held, then returned
const (
	callAlone        = 1
	heldThenReturned = 2
)

// onlyExpression is what a return of one value, or a statement of one expression, holds; nil for any other.
func onlyExpression(statement ast.Stmt) ast.Expr {
	switch typed := statement.(type) {
	case *ast.ReturnStmt:
		if len(typed.Results) == 1 {
			return typed.Results[0]
		}
	case *ast.ExprStmt:
		return typed.X
	}
	return nil
}

// heldAndReturned is the call whose results the first statement holds and the second returns, the same names in the
// same order; nil for any other pair.
func heldAndReturned(pass *analysis.Pass, first, second ast.Stmt) *ast.CallExpr {
	assigned, isAssigned := first.(*ast.AssignStmt)
	returned, isReturn := second.(*ast.ReturnStmt)
	if !isAssigned || !isReturn || len(assigned.Rhs) != 1 || len(assigned.Lhs) != len(returned.Results) {
		return nil
	}
	for index, held := range assigned.Lhs {
		heldName, isHeldName := held.(*ast.Ident)
		givenName, isGivenName := returned.Results[index].(*ast.Ident)
		if !isHeldName || !isGivenName || heldName.Name != givenName.Name {
			return nil
		}
	}
	return unchangedCall(pass, assigned.Rhs[0])
}

// unchangedCall is an expression as a call, not a conversion, where what is called and each argument stands
// unchanged; nil for any other.
func unchangedCall(pass *analysis.Pass, expression ast.Expr) *ast.CallExpr {
	call, isCall := ast.Unparen(expression).(*ast.CallExpr)
	if !isCall || pass.TypesInfo.Types[call.Fun].IsType() || !standsUnchanged(pass, call.Fun) {
		return nil
	}
	for _, argument := range call.Args {
		if !standsUnchanged(pass, argument) {
			return nil
		}
	}
	return call
}

// standsUnchanged tells whether an expression is a name, a field of one, a literal, or a generic function given its
// types, with no work done on it.
func standsUnchanged(pass *analysis.Pass, expression ast.Expr) bool {
	switch typed := ast.Unparen(expression).(type) {
	case *ast.Ident, *ast.BasicLit:
		return true
	case *ast.SelectorExpr:
		return standsUnchanged(pass, typed.X)
	case *ast.IndexExpr:
		return pass.TypesInfo.Types[typed.Index].IsType() && standsUnchanged(pass, typed.X)
	case *ast.IndexListExpr:
		return !slices.ContainsFunc(typed.Indices, func(index ast.Expr) bool { return !pass.TypesInfo.Types[index].IsType() }) &&
			standsUnchanged(pass, typed.X)
	}
	return false
}

// readsUnexportedField tells whether a call reads a field its package keeps unexported: the call is then the
// package's way to it, a getter.
func readsUnexportedField(pass *analysis.Pass, call *ast.CallExpr) bool {
	for node := range ast.Preorder(call) {
		selector, isSelector := node.(*ast.SelectorExpr)
		if !isSelector {
			continue
		}
		// a variable read through a selector from inside its package is a field: a package's variable would be
		// written bare
		if field, isVariable := pass.TypesInfo.ObjectOf(selector.Sel).(*types.Var); isVariable && !field.Exported() {
			return true
		}
	}
	return false
}
