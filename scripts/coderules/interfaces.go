package main

import (
	"go/types"
	"slices"
)

// isAskedFor tells whether a declared thing is a method of a type that implements an interface asking for it, its
// name and shape fixed by that interface. An interface's own method asks, and is not asked for.
func isAskedFor(declared types.Object, interfaces []*types.Interface) bool {
	method, isMethod := declared.(*types.Func)
	if !isMethod || method.Signature().Recv() == nil {
		return false
	}
	receiver := method.Signature().Recv().Type()
	if types.IsInterface(receiver) {
		return false
	}
	return slices.ContainsFunc(interfaces, func(asking *types.Interface) bool {
		return asksFor(asking, method.Name()) && (types.Implements(receiver, asking) || types.Implements(types.NewPointer(receiver), asking))
	})
}

func asksFor(asking *types.Interface, name string) bool {
	for method := range asking.Methods() {
		if method.Name() == name {
			return true
		}
	}
	return false
}

// interfacesInScope are the interfaces the package and the packages it imports declare, and error.
func interfacesInScope(checked *types.Package) []*types.Interface {
	interfaces := slices.Concat(interfacesIn(types.Universe), interfacesIn(checked.Scope()))
	for _, imported := range checked.Imports() {
		interfaces = append(interfaces, interfacesIn(imported.Scope())...)
	}
	return interfaces
}

// interfacesIn are the interfaces a scope declares.
func interfacesIn(scope *types.Scope) []*types.Interface {
	interfaces := []*types.Interface{}
	for _, name := range scope.Names() {
		typeName, isTypeName := scope.Lookup(name).(*types.TypeName)
		if !isTypeName {
			continue
		}
		if declared, isInterface := typeName.Type().Underlying().(*types.Interface); isInterface {
			interfaces = append(interfaces, declared)
		}
	}
	return interfaces
}
