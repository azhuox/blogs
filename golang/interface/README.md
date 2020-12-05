# Interfaces In Go

## Preface

Interface is a very important feature in Go as it is a key to implement polymorphism and dependency injection in Go. 
In spite of its importance, there are some serious arguments about the best practice of utilizing Golang, which 
may cause another [little endian v.s. big endian war](https://www.ling.upenn.edu/courses/Spring_2003/ling538/Lecnotes/ADfn1.htm) in the future. 
The purpose of this blog to is discuss these arguments in a peaceful way and try to eliminate the potential war (if possible).

## Introduction to Go Interface

The following shows an example of Go interface:

```
// Product defines the product we want to produce.
type Product struct {
}

// Producer defines the methods that a producer should implement.
type Producer interface {
    // Produce 
	Produce() *Product
	Type() string
}

// SimpleProducer is a simple producer that implements "Producer" interface.
type SimpleProducer struct {
    
}

func (p *SimpleProducer) Produce() *Product {
	return &Product {}
}

func (p *SimpleProducer) Type() string {
    return "Simple producer"   
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

**Accept Interfaces but Return Structs**

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

Normally, the smaller the interface，the stronger the abstraction. The stronger the abstraction, the easier to be accepted to return interfaces. 


**“Go interfaces generally belong in the package that uses values of the interface type, not the package that implements those values”**

The above sentence comes from [Go Code Review Comments, Interface Section](https://github.com/golang/go/wiki/CodeReviewComments#interfaces), which 
totally conflicts with what it is said in [Effective Go](https://golang.org/doc/effective_go.html#generality):

"If a type exists only to implement an interface and will never have exported methods beyond that interface, 
there is no need to export the type itself. Exporting just the interface makes it clear the value has no interesting behavior beyond what is described in the interface."

So where should we define interfaces? Consumer side or producer side? I think this depends on whether the interface that you define has good abstraction or not. 
If an interface has very strong abstraction, then it makes sense to put it in an individual shared package. 
The interface `io.Reader`  in the [package io](https://github.com/golang/go/blob/c170b14c2c1cfb2fd853a37add92a82fd6eb4318/src/io/io.go#L77-L92) is a
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






## Reference

- [Interface Segregation Principle](https://en.wikipedia.org/wiki/Interface_segregation_principle)
- [Effective Go](https://golang.org/doc/effective_go.html#generality)
- [package io](https://github.com/golang/go/blob/c170b14c2c1cfb2fd853a37add92a82fd6eb4318/src/io/io.go#L77-L92)
- [Go client of Github](https://github.com/google/go-github)



- introduction about interface
- Why I want to write this blog

## Interface in Go
- Behaviour contract
- Implicit implementation

## Why Interface

- Abstraction
- Dependency Injection.

## Arguments about Interface

### Return Structs or Return Interface?

### Define Interface in Consumer Side or Provider Side?

### Summary




 

## Use Cases of Go Interface

### Abstraction

General Principles:
- “The bigger the interface, the weaker the abstraction”

Good example

io package



### Dependency Injection

k8s package.
visitor

Dependency Injection sometimes does not mean Good Abastraction.

User package

break interface with muliple ones


## Arguments about Go Interface

### Where to Define Interface?

### Should Concrete Structs Be Exposed?



