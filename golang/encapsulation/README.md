# Encapsulation In Go

![Newton with A New Apple](https://github.com/azhuox/blogs/blob/master/golang/encapsulation/assets/newton-apple.jpg?raw=true)

## Preface

Encapsulation, as known as information hiding, is a key aspect of object-oriented programming. An object's field or method is said to be
encapsulated if it is inaccessible to users of the object. Unlike classical objected programming languages like Java, Go
has very specific encapsulation rules. This blog is going to explore these "interesting" rules.

## Encapsulation Rules in Go

Go has only one rule to set up encapsulation: capitalized identifiers are exported from the package where they are defined
and un-capitalized ones are not. **A field/method of a struct/interface is exported only when the names of the field/method and struct/interface
are both capitalized (AND condition).**

Let's go through an example to discuss the difference between Java and Go in terms of encapsulation.

Suppose we want to define a simple counter (without consideration of race condition), we can realize it in Go and Java in the following way:

```go
package encapsulationexample

// SimpleCounter - a simple counter.
type SimpleCounter struct {
   counter int
}

// NewCounter - creates a new counter instance.
func NewCounter() *SimpleCounter {
   return &SimpleCounter{
       counter: 0,
   }
}

// N - return current counter's value.
func (c *SimpleCounter) N() int { return c.counter }

// Increment - increases the counter by 1 without considering the risk of race condition.
func (c *SimpleCounter) Increment() { c.counter++ }

// Reset - resets the counter.
func (c *SimpleCounter) Reset() { c.counter = 0 }
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

In Java, `counter` field of the `Counter` class can only be accessed within the class due to the `private` directive. Clients of
`Counter` class can only access its public methods and fields. **In Go, `counter` field of `Counter` struct can be directly
accessed by other structs or functions defined in the same package**, like this:

```go
package encapsulationexample

import "fmt"

// CounterDemo shows how to directly access "private" fields of `Counter` struct in the same package.
func CounterDemo() {
  c := NewCounter()
  // I can directly access the `counter` field!
  c.counter = 5
  // It will print out "counter value: 5".
  fmt.Print("counter value: %d\n", c.N())
}

package otherpackage

import "fmt"

// CounterDemo shows "private" fields of `Counter` struct cannot be directly accessed by another package.
func CounterDemo2() {
  c := NewCounter()
  // This would cause a compiler error as the private field `counter` is not visible in this package.
  c.counter = 5
  // It will print out "counter value: 0" if you comment the above statement.
  fmt.Print("counter value: %d\n", c.N())
}
```

![Shocking Cat](https://github.com/azhuox/blogs/blob/master/golang/encapsulation/assets/shocking-cat.gif?raw=true)

**From the above example, you can see: unlike Java or other object-oriented programming languages that control visibility of names on class level,
Go controls encapsulation at package level.** The work around to fix this is to define a `Counter` interface and use it to replace usage of `SimpleCounter` struct. In this way, only the `SimpleCounter` struct can access its private fields and methods (see the following example).
In other words, Go interface can help you achieve information hiding on interface/struct level.

```go
package encapsulationexample

// Counter defines the interface that a counter instance needs to implement.
type Counter interface{
  N() int
  Increment()
  Reset()
}

// SimpleCounter - a simple counter.
type simpleCounter struct {
   counter int
}

func NewSimpleCounter() Counter {
   return &simpleCounter{
       counter: 0,
   }
}

func (c *simpleCounter) N() int { return c.counter }

func (c *simpleCounter) Increment() { c.counter++ }

func (c *simpleCounter) Reset() { c.counter = 0 }

// CounterDemo shows how to use the `Counter` interface.
func CounterDemo() {
  var c Counter
  c = NewSimpleCounter()
  // This would cause a compiler error as the `Counter` interface does not have `counter` field.
  c.counter = 5
  // It will print out "counter value: 0" if you comment the above statement.
  fmt.Print("counter value: %d\n", c.N())
}
```

## Encapsulation in Internal Packages

With above encapsulation rules, [Go internal packages](https://docs.google.com/document/d/1e8kOo3r51b2BWtTs_1uADIA5djfXhPT36s6eHVRIvaU/edit) have an extra rule:

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

- Packages defined in `foo/internal/` folder can be imported by packages defined in the directory rooted at `foo/` no matter how deep their directory layouts are. For example, `foo/cmd/server` package can import the `foo/internal/pkg/pkg2` package.
- The deepest `internal` dominates encapsulation rules when there are multiple `internals` in a package's import path. For example,
`foo/internal/module1/service/internal/repo` package can only be imported by packages in the directory tree rooted at `foo/internal/module1/service/` (other than `foo/`), which is only `foo/internal/module1/service` package in this case.


### When to Use Internal Packages

When to use internal packages? We only need to remember one rule: Define a package in the `internal` folder when you want it to be shared only among packages rooted at the parent of the “internal” directory. Take the above project layout of `foo` project (microservice) as
an example, there are two typical use cases of internal packages:

1. Define a project's internal packages: `foo/internal/` folder in the above example saves all the packages that can only be used in this project. This is because `foo/internal` is rooted at root folder of `foo` project and all the packages defined in this project are also rooted
at the root folder of the `foo` project. Therefore, any package defined in `foo` project can access packages defined in the `foo/internal` folder.
2. Define a package's exclusive packages: `foo/internal/module1/service/internal/repo` package can only be used by `foo/internal/module1/service` package according to rules of internal packages and this makes sense in terms of domain driven design. Normally a domain driven "module"
consists of three packages: `api-server`, `service` and `repository` and only `service` can access `repository`. In other words, `service` package should own the `repository` package regarding design pattern and code organization. Therefore, in the example,
it makes sense to define the `repository` package under the internal folder of `foo/internal/module1/service`  package. 

## Some Interesting Study cases

Mystery of Go encapsulation rules sometimes can make you confused and lost. For example, here are some "interesting" examples about Go encapsulation.

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
func (r *UserRepo) CreateUser(ctx context.Context, req CreateUserRequest) error {
  req.createdAt = time.Now().UTC()
  ...
}
```

In this example, `CreateUserRequest` struct allows `UserRepo` to control what to expose to users: When creating a user, a caller uses public fields `CreateUserRequest` struct to pass exposed parameters while `UserRepo` uses private fields of `CreateUserRequest` struct to set up internal parameters.
This prevents callers from setting some metadata that are exclusively controlled by `UserRepo`.

### Private Interface with Private Methods.

```go
package user
import "context"

// helper defines helper functions that are used in this package.
type helper interface {
  doSomething(ctx context.Context) error
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

You can define a private interface with some private methods for the purpose of **dependency injection** other than **abstraction**. For example, `helper` interface in the above example makes the `defaultHelper.doSomething` method replace-able by the `helpMock.doSomething` method.

Should a private interface own some public methods? No, it SHOULD NOT as public methods in a private interface will never get a chance to be exported. 

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

func (s *EventService) ProcessEvents(rawEvent *RawEvent) error {
  switch rawEvent.Type {
  case "meeting-ended":
     event := meetingEndedEvent{}
     _ = json.Unmarshal([]byte(rawEvent.String()), &event)
     ...
  }
}

```

A private data struct with public fields means it can only be used within the package where it is defined and those public fields are for marshal/unmarshal purpose. A field must be capitalized if it wants to be marshalled/unmarshalled.

### Private Object Structs with Public Methods.
```go
package io

// Reader is the interface that wraps the basic Read method.
type Reader interface {
  Read(p []byte) (n int, err error)
}

type multiReader struct {
  readers []Reader
}

func (mr *multiReader) Read(p []byte) (n int, err error) {
  // Blah blah blah.
  return 0, EOF
}

// MultiReader returns a Reader that's the logical concatenation of the provided input readers.
func MultiReader(readers ...Reader) Reader {
  r := make([]Reader, len(readers))
  copy(r, readers)
  return &multiReader{r}
}

```

A private object struct with public methods means it implements an public interface and has no interest exposing itself.
This follows [Generality principle](https://golang.org/doc/effective_go.html#generality) defined in [Effective Go](https://golang.org/doc/effective_go.html): "If a type exists only to implement an interface and will never have exported methods beyond that interface,
there is no need to export the type itself. Exporting just the interface makes it clear the value has no interesting behavior beyond what is described in the interface. It also avoids the need to repeat the documentation on every instance of a common method."

The above code is copied from built in [Go io package](https://golang.org/pkg/io/). You can see that the `multiReader` struct only exposes `Reader` interface.

## Summary

In summary, Go has following encapsulation rules:

- Go controls visibility of names at package level. A field/method of a struct/interface is exported only when the names of the field/method and struct/interface are both capitalized (AND condition).
- An import of a path containing the element “internal” is disallowed if the importing code is outside the tree rooted at the parent of the “internal” directory.
- A field must be capitalized if it wants to be JSON marshal/unmarshal, no matter whether the struct it belongs to is capitalized or not.


## Reference

- [The Go Programming Language, Chapter 6, Section 6: Encapsulation](https://github.com/KeKe-Li/book/blob/master/Go/The.Go.Programming.Language.pdf)
- [Design Document of Go Internal Package](https://docs.google.com/document/d/1e8kOo3r51b2BWtTs_1uADIA5djfXhPT36s6eHVRIvaU/edit)
- [Effective Go](https://golang.org/doc/effective_go.html)
- [Generality Principle About Interface and Encapsulation](https://golang.org/doc/effective_go.html#generality)
- [Go io package](https://golang.org/pkg/io/)
