package main

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/types/typeutil"
)

const (
	namedHelperMessage = "a function given a name inside another is a helper: make it a function of the package"
	longHelperMessage  = "a helper with logic is a function of the package, not a literal inside another"
)

const noNestedHelpersDescription = `refuse a helper written inside a function

A function literal inside a function is a lambda when it is handed straight to a call, returned or set as a
field, and is one simple statement: a comparison handed to a sort, a close handed back as a teardown, an
error joined in a defer, a result sent from a goroutine. One with more is a helper, and so is any literal
given a name, however short: a helper is a function of its package, with a name and parameters of its own.
The body handed to a test's or a benchmark's Run, a fuzz test's Fuzz or synctest's Test is the test itself,
and the iterator a function returns is the value it makes: both are let alone, and what is inside them is
held to the rule again.`

// What parts a package's path from a name in it, as the iterators are written
const packageSeparator = "."

// literalRole is what a function literal is to the code round it, known before the walk reaches it.
type literalRole int

const (
	// handedOver is a literal passed to a call, returned or set as a field: a lambda if it is one statement
	handedOver literalRole = iota
	// testBody is the body handed to a function that runs a test
	testBody
	// madeValue is the iterator a function returns as its result
	madeValue
	// named is a literal given a name by an assignment or a declaration
	named
)

// helperRule is what the rule reads a call and a result against, each table by full name.
type helperRule struct {
	// testRunners are the functions whose literal argument is a test's body
	testRunners map[string]bool
	// iterators are the types of a function's result whose literal is the value the function makes
	iterators map[string]bool
}

func noNestedHelpersAnalyzer() *analysis.Analyzer {
	rule := helperRule{testRunners: setOf(testRunners()), iterators: setOf(iterators())}
	return &analysis.Analyzer{
		Name: "nonestedhelpers",
		Doc:  noNestedHelpersDescription,
		Run:  func(pass *analysis.Pass) (any, error) { return runNoNestedHelpers(pass, rule) },
	}
}

// testRunners are the functions whose literal argument is a test's body.
func testRunners() []string {
	return []string{"(*testing.T).Run", "(*testing.B).Run", "(*testing.F).Fuzz", "testing/synctest.Test"}
}

// iterators are the types of a function's result whose literal is the value the function makes.
func iterators() []string {
	return []string{"iter.Seq", "iter.Seq2"}
}

func runNoNestedHelpers(pass *analysis.Pass, rule helperRule) (any, error) {
	for _, file := range writtenFiles(pass) {
		for _, declaration := range file.Decls {
			function, isFunction := declaration.(*ast.FuncDecl)
			if isFunction && function.Body != nil {
				reportHelpersIn(pass, rule, function)
			}
		}
	}
	return nil, nil
}

// reportHelpersIn walks a function's body. A call, an assignment and a declaration are met before the
// literals they hold, so each literal's role is known by the time the walk reaches it.
func reportHelpersIn(pass *analysis.Pass, rule helperRule, function *ast.FuncDecl) {
	roles := map[*ast.FuncLit]literalRole{}
	if returnsAnIterator(pass, rule, function) {
		markReturnedLiterals(function.Body, roles)
	}
	ast.Inspect(function.Body, func(node ast.Node) bool { return visitForHelpers(pass, rule, node, roles) })
}

func visitForHelpers(pass *analysis.Pass, rule helperRule, node ast.Node, roles map[*ast.FuncLit]literalRole) bool {
	switch typed := node.(type) {
	case *ast.CallExpr:
		if runsATest(pass, rule, typed) {
			markLiterals(typed.Args, testBody, roles)
		}
	case *ast.AssignStmt:
		markNamedLiterals(typed, roles)
	case *ast.ValueSpec:
		markLiterals(typed.Values, named, roles)
	case *ast.FuncLit:
		reportHelper(pass, typed, roles[typed])
	}
	return true
}

func markLiterals(expressions []ast.Expr, role literalRole, roles map[*ast.FuncLit]literalRole) {
	for _, expression := range expressions {
		if literal, isLiteral := expression.(*ast.FuncLit); isLiteral {
			roles[literal] = role
		}
	}
}

// markNamedLiterals marks a literal assigned to a name; one assigned to a field or an element is set as a
// field, and stays handed over.
func markNamedLiterals(assignment *ast.AssignStmt, roles map[*ast.FuncLit]literalRole) {
	if len(assignment.Lhs) != len(assignment.Rhs) {
		return
	}
	for index, target := range assignment.Lhs {
		literal, isLiteral := assignment.Rhs[index].(*ast.FuncLit)
		if _, isName := target.(*ast.Ident); isLiteral && isName {
			roles[literal] = named
		}
	}
}

func markReturnedLiterals(body *ast.BlockStmt, roles map[*ast.FuncLit]literalRole) {
	for _, statement := range body.List {
		if returned, isReturn := statement.(*ast.ReturnStmt); isReturn {
			markLiterals(returned.Results, madeValue, roles)
		}
	}
}

func reportHelper(pass *analysis.Pass, literal *ast.FuncLit, role literalRole) {
	switch role {
	case testBody, madeValue:
		// let alone; what is inside it is walked and held to the rule
	case named:
		pass.Reportf(literal.Pos(), namedHelperMessage)
	case handedOver:
		if !isOneStatement(literal) {
			pass.Reportf(literal.Pos(), longHelperMessage)
		}
	}
}

// runsATest tells a call to one of the testRunners, by the function it calls.
func runsATest(pass *analysis.Pass, rule helperRule, call *ast.CallExpr) bool {
	called, isFunction := typeutil.Callee(pass.TypesInfo, call).(*types.Func)
	return isFunction && rule.testRunners[called.FullName()]
}

// returnsAnIterator tells a function whose one result is one of the iterators.
func returnsAnIterator(pass *analysis.Pass, rule helperRule, function *ast.FuncDecl) bool {
	declared, isFunction := pass.TypesInfo.Defs[function.Name].(*types.Func)
	if !isFunction {
		return false
	}
	results := declared.Signature().Results()
	if results.Len() != 1 {
		return false
	}
	result, isNamed := results.At(0).Type().(*types.Named)
	return isNamed && result.Obj().Pkg() != nil &&
		rule.iterators[result.Obj().Pkg().Path()+packageSeparator+result.Obj().Name()]
}

// isOneStatement tells a lambda: a body of one simple statement, an expression, an assignment, a send, a step
// of a counter or a return.
func isOneStatement(literal *ast.FuncLit) bool {
	if len(literal.Body.List) != 1 {
		return false
	}
	switch literal.Body.List[0].(type) {
	case *ast.ExprStmt, *ast.AssignStmt, *ast.SendStmt, *ast.IncDecStmt, *ast.ReturnStmt:
		return true
	default:
		return false
	}
}
