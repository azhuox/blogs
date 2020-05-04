# Encapsulation In Go

## Preface

Encapsulation, as known as information hiding, is a key aspect of objected-oriented programming. An object's field or method is said to be
encapsulated if it is inaccessible to users of the object. However, unlike classical objected programming languages, for example, Java, Go
has very specific rules about encapsulation.


## Encapsulation Rules in Go

Go has only one rule to set up encapsulation: capitalized identifiers are exported from the package where they are defined
and un-capitalized ones are not. A field or method of a struct is exported only when the names of the field/method and struct
are both capitalized (AND condition).

Let's go through an example to the difference between Java and Go in terms of encapsulation.

Suppose we want to define a simple counter without consideration of race condition, here is how Jave code looks like:

```go
package encapsulationexample

// Counter - a simpler counter.
type Counter struct {
    counter int
}

// NewCounter - creates a new counter instance.
func NewCounter() *Counter {
    return &Counter{
        counter: 0,
    }
}

// N - return current counter's value.
func (c *Counter) N() int { return c.counter }

// Increment - increases the counter by 1 without considering the risk of race condition.
func (c *Counter) Increment() { c.counter++ }

// Reset - resets the counter.
func (c *Counter) Reset() { c.counter = 0 }
```

```java
public class Counter {
    // counter variable, only visible in the class.
    private int counter;

    public Counter() {
        this.counter = 0;
    }

     // N - return current counter's value.
     public int N() {
        return this.counter;
     }

     // Other methods are omitted.
     ......
}
```

In Java, `counter` field of `Counter` class can only be accessed within the class due to `private` directive. Clients of
`Counter` class can only access its public methods and fields. In Go, `counter` field of `Counter` struct can be directly
accessed by other structs or functions defined in the same package, like this:

```go
package encapsulationexample

import "fmt"

// CounterDemo shows how to directly access "private" field of `Counter` struct.
func CounterDemo() {
	c := NewCounter()
	// I can directly access `counter` field!
	c.counter = 5
	// It will print out "counter value: 5".
	fmt.Print("counter value: %d\n", c.N())
}

package otherpackage

import "fmt"

// CounterDemo shows "private" field of `Counter` struct cannot be directly accessed by another package.
func CounterDemo2() {
	c := NewCounter()
	// This won't work! It would cause a compiler error.
	c.counter = 5
	// It will print out "counter value: 0".
	fmt.Print("counter value: %d\n", c.N())
}
```

[intertering gif]

**From the above example, you can see: unlike Java or other object-oriented programming languages that controls visibility of names on class level, 
Go sets up encapsulation on package level.** The work around to fix this is to define a `Counter` interface:

```go


```




Unlike other object-oriented programming languages like Java, Go has very specific rules about encpsulaiton:
1. based on package

Example:
    In java
    In Go

2. Internal package

- Use Cases of Go Encapsulation

- Public Data Structs

- Private Data Structs

- Public Object Structs

- Private Object Structs

## Reference

- [The Go Programming Language, Chapter 6, Section 6: Encapsulation](https://github.com/KeKe-Li/book/blob/master/Go/The.Go.Programming.Language.pdf)
- [Design Document of Go Internal Package](https://docs.google.com/document/d/1e8kOo3r51b2BWtTs_1uADIA5djfXhPT36s6eHVRIvaU/edit)

