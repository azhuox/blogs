# Interfaces In Go

![Big Endian v.s. Little Endian](https://github.com/azhuox/blogs/blob/master/golang/interface/assets/little-end-big-end.jpg?raw=true)

## Preface

Interfaces is a very important feature in Go. It is a key to implement polymorphism and dependency injection in Go. 
In spite of its importance, there are some serious arguments about the best practice of Go interfaces, which 
may lead to another [little endian v.s. big endian war](https://www.ling.upenn.edu/courses/Spring_2003/ling538/Lecnotes/ADfn1.htm) in the future. 
The purpose of this blog to is discuss these arguments in a peaceful way and find a way to avoid this potential war (if possible).

## Introduction to Go Interfaces

The following shows an example of Go interface:

```go
package interfaceexample

import (
	"fmt"
)

// FlyBehaviour defines an interface for fly behaviour
type FlyBehaviour interface {
	Fly()
	Type() string
}

// FlyWithWings defines the fly behavior with wings.
type FlyWithWings struct {
}

func (f *FlyWithWings) Fly() {
	fmt.Print("Fly with wings!\n")
}

func (f *FlyWithWings) Type() string {
	return "fly-with-wings"
}

// FlyWithWings defines the fly behavior with super power. 
type FlyWithSuperPower struct {
}

func (f *FlyWithSuperPower) Fly() {
	fmt.Print("Fly with super power!\n")
}

func (f *FlyWithSuperPower) Type() string {
	return "fly-with-super-power"
}
```

From the above example, you can see that:
- A Go interface defines one or more methods.
- Unlike other languages, you don’t have to explicitly declare that a type implements an interface. 
A struct `S` is a thing which is defined by an interface `I` as long as the struct `S` implements all the methods defined by the interface `I`. 
In this example, `FlyWithWings` and `FlyWithSuperPower` are both `FlyBehaviour` as they implements all the methods defined in the interface `FlyBehaviour`. 

## Best Practice (a.k.a Arguments)

The concept and usage of Go interfaces is simple. However, there are some arguments regarding its best practice.
Now let us discuss those arguments and find some agreement regarding the best practice of Go interfaces.


**"The smaller the interface，the stronger the abstraction"**

There is no argument about this best practice as this is basically another expression of [Interface Segregation Principle](https://en.wikipedia.org/wiki/Interface_segregation_principle).

The following shows a example that follows this principle, in which the interface `user.Manager`  is split into multiple smaller interfaces. 

```go
package user
import (
    "context"
)

// Interface
//
type Getter interface{
	Get(ctx context.Context, ids []string) ([]*Users, error)
}

type Lister interface{
	List(ctx context.Context, options *ListOptions) ([]*Users, error)
}

type Reader interface{
	Getter
	Lister
}

type Creater interface{
	Creat(ctx context.Context, req *CreateRequest) (*User, error)
}

type Updater interface{
	Update(ctx context.Context, req *UpdateRequest) error
}

type Writer interface{
	Creater
	Updater
}

type ManagerInterface interface{
	Reader
	Writer
}

// Implementation

type Manager struct{
	
}

```

**"Accept Interfaces but Return Structs"**

We all agree on accepting interfaces other than concrete implementations as this is the key to follow [SOLID principles](https://en.wikipedia.org/wiki/SOLID) (e.g. [Dependency Inversion Principle](https://en.wikipedia.org/wiki/Dependency_inversion_principle))
and realize some design patterns (e.g. [Decorator Pattern](https://en.wikipedia.org/wiki/Decorator_pattern)) in Go. However, there is no solid answer to returning interfaces (abstraction) or structs (implementation). 

The benefit of returning structs is to give consumers freedom to define interfaces on their side. Take the above interface `user.ManagerInterface` as an example, 
consumers can define this interface based on the struct `user.Manager` (the implementation of user manager) and their need in their own packages. 

The benefit of returning interfaces it to free consumers from defining interfaces on their own, as long as the returning interfaces provide strong abstraction. Moreover, some design patterns,
such as [Factory Method Pattern](https://en.wikipedia.org/wiki/Factory_method_pattern), require to return interfaces other than concrete implementation. For example, the method [aes.newCipher](https://github.com/golang/go/blob/dcd3b2c173b77d93be1c391e3b5f932e0779fb1f/src/crypto/aes/cipher_asm.go#L33-L54)
returns the interface [cipher.Block](https://github.com/golang/go/blob/dcd3b2c173b77d93be1c391e3b5f932e0779fb1f/src/crypto/cipher/cipher.go#L15) with different structs in different circumstances:

```go
func newCipher(key []byte) (cipher.Block, error) {
	if !supportsAES {
		return newCipherGeneric(key)
	}
	n := len(key) + 28
	c := aesCipherAsm{aesCipher{make([]uint32, n), make([]uint32, n)}}
    ...

	if supportsAES && supportsGFMUL {
		return &aesCipherGCM{c}, nil
	}
	return &c, nil
}
```

Normally, the smaller the interface，the stronger the abstraction. The stronger the abstraction, the more acceptable it is to return interfaces. 


**“Go interfaces generally belong in the package that uses values of the interface type, not the package that implements those values”**

The above sentence comes from [Go Code Review Comments, Interface Section](https://github.com/golang/go/wiki/CodeReviewComments#interfaces), which 
totally conflicts with what it is said in [Effective Go](https://golang.org/doc/effective_go.html#generality):

"If a type exists only to implement an interface and will never have exported methods beyond that interface, 
there is no need to export the type itself. Exporting just the interface makes it clear the value has no interesting behavior beyond what is described in the interface."

So where should we define interfaces? Consumer side or producer side? I think this depends on whether the interface that you define has good abstraction or not. 
If an interface provides very strong abstraction, then it makes sense to put it in an individual shared package. 
The interface `io.Reader` in the [package io](https://github.com/golang/go/blob/c170b14c2c1cfb2fd853a37add92a82fd6eb4318/src/io/io.go#L77-L92) is a
perfect example of interfaces with strong abstraction. It only has one method `Read` and everyone agrees what this method should look like.
Therefore, everyone is fine with putting the interface `Reader` in the package `io`, which is a producer that provides implementations of
this interface.

```
// Reader is the interface that wraps the basic Read method.
// ......
type Reader interface {
	Read(p []byte) (n int, err error)
}
```

The opposite example of putting Go interfaces at the consumer side is [Go client of Github](https://github.com/google/go-github).
It is apparently not a good idea to define a `github.Client` interface as a Github client needs to provide a lot of methods, which 
leads to weak abstraction for making a `github.Client` interface. In this case, it makes more sense for consumers to define their 
own `github.Client` interfaces, which may vary from consumer to consumer and may only have very few methods. 

**Define a default abstraction for an implementation may be useful.**

It is easy to determine where to put an interface when the interface provides either very strong or very week abstraction. 
But we are not always that lucky in reality. Take the above interface `user.ManagerInterface` as an example, it makes sense
to split it into multiple smaller interfaces. But do these interfaces provide strong abstraction so that it is acceptable
to define them at the producer side (the package `user`)? Or should this producer only provide concrete implementation of user manager?

Suppose the package `user` is an internal package for a project and it has five consumers (five packages are using this package).
The interface `user.ManagerInterface` needs to be defined five times in these packages for programming to interfaces, 
if there is no default interface `user.ManagerInterface` and it is very likely these packages may define the same interface `user.ManagerInterface`.
Therefore, it can be very useful for the package `user` (producer) to provide a default interface and its mocks in this case. Here is the pseudo code: 

```go

// Implementation 

// Interface of user manager - optional
//
package user
import (
    "context"
)

type Getter interface{
	Get(ctx context.Context, ids []string) ([]*Users, error)
}

type Lister interface{
	List(ctx context.Context, options *ListOptions) ([]*Users, error)
}

type Reader interface{
	Getter
	Lister
}

type Creater interface{
	Creat(ctx context.Context, req *CreateRequest) (*User, error)
}

type Updater interface{
	Update(ctx context.Context, req *UpdateRequest) error
}

type Writer interface{
	Creater
	Updater
}

type ManagerInterface interface{
	Reader
	Writer
}

// Implementation of user manager - required
//
type Manager struct{
	...
}

func NewManager() *Manager {
	return &Manager{}
}

// Mock of user manager - optional
//
type ManagerMock struct{
	...
}
```     

From the above example, you can see that:
* The implementation `user.Manager` is exposed and returned, which gives consumers freedom to define the abstraction of `user.ManagerInterface`
base on this implementation.
* The package `user` provides a default abstraction of `user.Manager` and its mocks through the interface `user.ManagerInterface` and the strcut `user.ManagerMock`,
which frees some consumers from defining the abstraction of `user.Manager` and its mocks.
* This pattern is a trade off of the argument about where to define interfaces. That is, provide default abstraction at the producer side but also allow
consumers to define their own abstraction. And most importantly, let consumers decide what to use.  

## Summary

This blog talks about several aspects about interfaces, including the size of an interfaces, where to define interfaces,
return strcuts or interfaces, and the benefit of providing a default abstraction for an implementation. Like uncertainty of 
life, there is no solid answer to these arguments/questions. But I do hope this blog help you clean up some puzzle about Golang interfaces. 
And most importantly, may the world piece forever.

## Reference

- [Interface Segregation Principle](https://en.wikipedia.org/wiki/Interface_segregation_principle)
- [Effective Go](https://golang.org/doc/effective_go.html#generality)
- [package io](https://github.com/golang/go/blob/c170b14c2c1cfb2fd853a37add92a82fd6eb4318/src/io/io.go#L77-L92)
- [Go client of Github](https://github.com/google/go-github)

