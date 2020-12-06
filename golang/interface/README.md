# Interfaces In Go

## Preface

Interface is a very important feature in Go as it is a key to implement polymorphism and dependency injection in Go. 
In spite of its importance, there are some serious arguments about the best practice of utilizing Golang, which 
may cause another [little endian v.s. big endian war](https://www.ling.upenn.edu/courses/Spring_2003/ling538/Lecnotes/ADfn1.htm) in the future. 
The purpose of this blog to is discuss these arguments in a peaceful way and try to eliminate the potential war (if possible).

## Introduction to Go Interface

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
- A Go interface defines (but not implement) one or more methods.
- Unlike other languages, you don’t have to explicitly declare that a type implements an interface. 
A struct `S` is a thing which is defined by an interface `I` as long as the struct `S` implements all the methods that are defined by the interface `I`. 
In this example, the struct `SimpleProducer` is a `Producer` (defined by the interface `Producer`) as it implements all the methods that are defined in the interface `Producer`. 

## Best Practice (a.k.a Arguments)

The concept and usage of Go interface is simple. However, there are some arguments regarding the best practice of Go interfaces. Now let us those arguments and find the potential best practice.


**"The smaller the interface，the stronger the abstraction"**

There is no argument about this best practice as this is basically another expression of [Interface Segregation Principle](https://en.wikipedia.org/wiki/Interface_segregation_principle).

The following shows an interface example that follows this principle, in which a giant `user.Manager` interface is split into multiple smaller interfaces. 

```go
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

type Manager interface{
	Reader
	Writer
}
```

**"Accept Interfaces but Return Structs"**

We all agree on accepting interfaces other than concrete implementation as it is the key to follow [SOLID principles](https://en.wikipedia.org/wiki/SOLID) (e.g. [Dependency Inversion Principle](https://en.wikipedia.org/wiki/Dependency_inversion_principle))
and realize some design patterns (e.g. [Decorator Pattern](https://en.wikipedia.org/wiki/Decorator_pattern)) in Go. However, there is no solid answer about returning interfaces or structs. 

The benefit of returning structs is to give consumers freedom to define interfaces on their side. Take the above `user.Manager` as an example, 
consumers can define the interface `user.Manager` based on their need in their own packages if we define `user.Manager` as a struct other than an interface in the package `user`. 

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
If an interface has very strong abstraction, then it makes sense to put it in an individual shared package. 
The interface `io.Reader` in the [package io](https://github.com/golang/go/blob/c170b14c2c1cfb2fd853a37add92a82fd6eb4318/src/io/io.go#L77-L92) is a
perfect example of interfaces with strong abstraction. It only has one method `Read` and everyone agrees what the method `Read` should looks like in 
the official `io` package. Therefore, everyone is fine with putting this interface in the package `io`, which is a producer that provides the structs that implement
the interface `Reader`. 

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

**Define a default abstraction for an implementation may benefit 80% of users of that implementation.**

It is easy to determine where to put an interface when the interface provides either very strong or very week abstraction. 
But we are not always that lucky to easily in reality. That the above interface `user.Manager` as an example, it makes sense
to split it into multiple smaller interfaces. But do these interfaces provide strong abstraction so that it is acceptable
to define them at the producer side (the package `user`)? 

Suppose the interface `user.Manager` has five consumers. Let us find out what they need before determining where to define this interface.

Consumer A: I want a `user.Manager` interface that allows me to get users by user IDs.

Consumer B: I want a `user.Manager` interface that allows me to get a page of users ascended by create time of these users. 

Consumer C: I want a `user.Manager` interface that allows me to create users and get users by user IDs.

Consumer D: I want a `user.Manager` interface that allows me to modify user information.
    
Consumer E: I want a `user.Manager` interface that allows me to create a user if the user does not exist, or update a user if the user exists. 

Producer (the package `user`): OK, I know what you need now. How about this: I can provide you a `user.Manager` implementation 
which has all the CRUD features that you need. Moreover, I will provide you `user.ManagerInterface` interface which is composed by 
multiple small interfaces and its mocks. This is my version of `user.Manager` interface, feel free to use it 
(and its mocks for writing unit tests) if it covers your need. Or you can create your own `user.Manager` interface based on
the implementation I provide you. Here is the pseudo code:

```go

// Implementation 

// Interface
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

// Implementation
//
type Manager struct{
	...
}

// Mock
//
type ManagerMock struct{
	...
}
```

Consumer A, B, C, D: Great! Thank you for defining the interface `user.ManagerInterface` and its mocks, which covers
our need and saves us a lot of time defining them!

Consumer E: Thanks for your effort! However, I am going to define my own `user.ManagerInterface` interface based on 
the implementation you provide as the abstraction you provide does not cover my need.       

From the above example, the interface `user.ManagerInterface` that producer provides is more like a default interface (abstraction)
for the struct `user.Manager` and 80% of consumers may benefit from it. Moreover, exposing `user.Manager` allows consumers
to define their own `user.ManagerInterface` abstraction. In other words, it defines a default abstraction for users but 
also gives users freedom to define their own abstraction about `user.Manager`.  

## Summary

This blog talks about several aspects about interfaces, including the size of an interfaces, where to define interfaces,
return strcuts or interfaces, and the benefit of providing default abstraction for an implementation. Like uncertainty of 
life, there is no solid answer to these arguments/questions. But I do hope this blog help you clean up some puzzle about Golang interfaces. 
And most importantly, may the world piece forever.

## Reference

- [Interface Segregation Principle](https://en.wikipedia.org/wiki/Interface_segregation_principle)
- [Effective Go](https://golang.org/doc/effective_go.html#generality)
- [package io](https://github.com/golang/go/blob/c170b14c2c1cfb2fd853a37add92a82fd6eb4318/src/io/io.go#L77-L92)
- [Go client of Github](https://github.com/google/go-github)

