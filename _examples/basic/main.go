package main

import (
	"fmt"
	"github.com/hackborn/sqi"
)

func main() {
	// User-defined data model
	type Person struct {
		Name     string
		Children []Person
	}

	// User-defined data
	parent := &Person{"Katie", []Person{
		Person{"Eleanor", nil},
		Person{"Jason", nil},
	}}

	// Get a single top-level value
	name := sqi.EvalString("/Name", parent, nil)

	// Get a collection of nested values.
	children, _ := sqi.Eval("/Children", parent, nil)

	// Get a value from a child.
	childname := sqi.EvalString("/Children[0]/Name", parent, nil)

	fmt.Println("Parent's name is", name)
	fmt.Println("Parent's children are", children)
	fmt.Println("First child's name is", childname)
}
