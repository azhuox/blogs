# Encapsulation In Go

## Preface

Encapsulation, as known as information hiding, is a key aspect of objected-oriented programming. An object's field or method is said to be
encapsulated if it is inaccessible to users of the object. However, unlike classical objected programming languages, for example, Java, Go
has very specific rules about encapsulation.


## Encapsulation Rules in Go

Go has only one rule to set up encapsulation: capitalized identifiers are exported from the package where they are defined
and un-capitalized ones are not. **A field/method of a struct/interface is exported only when the names of the field/method and struct/interface
are both capitalized (AND condition).**

Let's go through an example to the difference between Java and Go in terms of encapsulation.

Suppose we want to define a simple counter without consideration of race condition, here is how Jave code looks like:

```go
package encapsulationexample

// Counter - a simple counter.
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
	// This would cause a compiler error as private field `counter` is not visible in this package.
	c.counter = 5
	// It will print out "counter value: 0" if you comment the above statement.
	fmt.Print("counter value: %d\n", c.N())
}
```

[intertering gif]

**From the above example, you can see: unlike Java or other object-oriented programming languages that controls visibility of names on class level, 
Go sets up encapsulation on package level.** The work around to fix this is to define a `Counter` interface and use it everywhere. In other words,
Go interface can help you achieve information hiding on interface/struct level.

```go
package encapsulationexample

// Counter defines interface that a counter instance needs to implement.
type Counter interface{
	N() int
	Increment()
	Reset()
}

// SimpleCounter - a simple counter.
type SimpleCounter struct {
    counter int
}

func NewSimpleCounter() *SimpleCounter {
    return &SimpleCounter{
        counter: 0,
    }
}

func (c *SimpleCounter) N() int { return c.counter }

func (c *SimpleCounter) Increment() { c.counter++ }

func (c *SimpleCounter) Reset() { c.counter = 0 }

// CounterDemo shows how to use `Counter` interface.
func CounterDemo() {
	var c Counter
	c := NewSimpleCounter()
	// This would cause a compiler error as `Counter` interface does not have `counter` field.
	c.counter = 5
	// It will print out "counter value: 0" if you comment the above statement.
	fmt.Print("counter value: %d\n", c.N())
}
```


## Encapsulation in Internal Packages

The above encapsulation rules can also be applied to [Go internal packages](https://docs.google.com/document/d/1e8kOo3r51b2BWtTs_1uADIA5djfXhPT36s6eHVRIvaU/edit). In addition, the following rule is also adopted in internal packages:

- "An import of a path containing the element “internal” is disallowed if the importing code is outside the tree rooted at the parent of the “internal” directory." - [Design Document of Go Internal Package](https://docs.google.com/document/d/1e8kOo3r51b2BWtTs_1uADIA5djfXhPT36s6eHVRIvaU/edit)

Here is an example:

```
foo:    -> repo 
    cmd:
        server:
            main.go
    internal:
        module:
            module1:
                service:
                    service.go
                    internal:
                        repo:
                            repo.go
        pkg:
            pkg1:
                code.go
            pkg2:
                code.go
    pkg:
        pkg1:
            code.go
```

In this case:

- Packages in `foo/internal/*` directory can be imported by packages in the directory rooted at `foo/` no matter how deep their directory layouts are. For example, `food/cmd/server` package can import `foo/internal/pkg/pkg2` package.
- The deepest `internal` dominates encapsulation rules when there are multiple `internals` in a package's import path. For example,
 `foo/internal/module1/service/internal/repo` package can only be imported by packages in the directory tree rooted at `foo/internal/module1/service/` (other than `foo/`), which only is `foo/internal/module1/service` package in this case. 


## Use cases

The following are some "interesting" examples about Go encapsulation.

### Public Data Structs with Private Fields.

```go
package user
import (
    "context"
    "time"

)

// CreateUserRequest - request body for creating a user.
type CreateUserRequest struct {
	Name string
	Email string
	createdAt time.Time
}

// CreateUser - creates a user in the database.
func (r *repo) CreateUser(ctx context.Context, req CreateUserRequest) error {
	req.createdAt = time.Now().UTC()
	...
}
```

### Private Interface with Private Methods.

```go
package user
import "context"

// helper defines helper functions that are used in this package.
type helper interface {
	func doSomething(ctx context.Context) error
}

type helperDefault struct {}

func (h *defaultHelper) doSomething(ctx context.Context) error {
	...
	return nil
}

type helperMock struct {}

// doSomething - mock of `doSomething` method.
func (h *helperMock) doSomething(ctx context.Context) error {
	...
	return nil
}
```


### Private Data Structs with Public Fields.
```go
package meetingevents

import "encoding/json"

type meetingEndedEvent struct {
	MeetingID string `json:"meeting_id"`
	HostID string `json:"host_id"`
	Duration int `json:"duration"`
	...
}

func (s *SVCDefault) ProcessEvents(rawEvent *RawEvent) error {
	switch rawEvent.Type {
	case "meeting-ended":
		event := meetingEndedEvent{}
		_ = json.Unmarshal([]byte(rawEvent.String()), &event)
		...
	}
}

```

### Private Object Structs with Public Fields.


## Summary

In summary, Go has following encapsulation rules:

- Go controls visibility of names at package level. A field/method of a struct/interface is exported only when the names of the field/method and struct/interface
are both capitalized (AND condition).
- An import of a path containing the element “internal” is disallowed if the importing code is outside the tree rooted at the parent of the “internal” directory.
- A field must be capitalized if it wants to be JSON marshal/unmarshal, not matter whether the struct it belongs to is capitalized or not.


## Reference

- [The Go Programming Language, Chapter 6, Section 6: Encapsulation](https://github.com/KeKe-Li/book/blob/master/Go/The.Go.Programming.Language.pdf)
- [Design Document of Go Internal Package](https://docs.google.com/document/d/1e8kOo3r51b2BWtTs_1uADIA5djfXhPT36s6eHVRIvaU/edit)

