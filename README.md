# sqi
A simple Go query library for interface{}s.

**Sqi** lets you ask questions about your data. It can be used to find values and extract results from complex user data, as well as data that has been hydrated via an encoding such as JSON.

## QUICK START ##

Sqi accesses nested data via a simple topic string. It provides the general function `Eval()` for locating a result, along with conveniences such as `EvalInt()`, `EvalString()`, etc. for typed results. The general use case is to provide a path (`/`) delimited string and data to an Eval() function.

Here's a complete example:


```
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
```

The output is
```
Parent's name is Katie
Parent's children are [{Eleanor []} {Jason []}]
First child's name is Eleanor
```


## CREDIT ##

Much thanks to a couple people who have provided great info on top down operator precedence parsers:
[Cristian Dima](http://www.cristiandima.com/top-down-operator-precedence-parsing-in-go)  
[Eli Bendersky](https://eli.thegreenplace.net/2010/01/02/top-down-operator-precedence-parsing)