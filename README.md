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

## OPERATORS ##

Examples for the following operators use this data model:

```
type Person struct {
	Name     string
	Age      int
	Children []Person
}
```

### PATH ###

The path `/` operator answers the field or map key with the given value.

Example:
```
sqi.EvalString(`/Name`, &Person{Name: "Ana"})
```
results in `"Ana"`.

### EQUALS ###

The equals `==` operator answers true if the left side matches the right side, false otherwise.

Example:
```
sqi.EvalBool(`/Name == 22`, &Person{Age: 22})
```
results in `true`.

### NOT EQUALS ###

The not equals `!=` operator is the opposite of equals.

### AND ###

The and `&&` operator evalutes to true if both the left and right sides are true.

### OR ###

The or `||` operator evalutes to true if either the left or right side is true.

### PARENTHESES ###

The parentheses `()` operator encapsulates a phrase.

### ARRAY ###

The array `[]` operator answers an element of an array or slice. Currently it only supports a single index value.

Example:
```
sqi.Eval(`/Children[0]`, &Person{Children: []Person{Person{Name: "a"}}})
```
results in `Person{Name: a}`.

## TECHNIQUES ##

### SELECT ###

A select finds one or more items from a collection. Selects in sqi are basically paths with an equality statement.

Example 1. Answer a collection of one or more items.
```
sqi.Eval(`/Children/(/Name == "c")`, &Person{Children: []Person{Person{Name: "a"}, Person{Name: "c"}}})
```
results in `[]Person{Person{Name: c}}`.

Example 2. Answer a single value from a collection.
```
sqi.Eval(`(/Children/(/Name == "c"))[0]`, &Person{Children: []Person{Person{Name: "a"}, Person{Name: "c"}}})
```
results in `Person{Name: c}`.

## CREDIT ##

Much thanks to a couple people who have provided great info on top down operator precedence parsers:
[Cristian Dima](http://www.cristiandima.com/top-down-operator-precedence-parsing-in-go)\
[Eli Bendersky](https://eli.thegreenplace.net/2010/01/02/top-down-operator-precedence-parsing)