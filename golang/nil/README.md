## What is `nil` in Go
`nil` in Go has several meanings:
- It represents "null" in Go. This means two things: 1. It does not have type. 2. Its value is "null".
- It is a predeclared identifier in Go, which means you can use it without declaring it.
- It represents zero values (and default values) of some types in Go, including:
    - interface types
    - pointer types
    - slice types
    - map types
    - channel types
    - function types


## Using `nil` as Zero Values

`nil` represents zero values (and default values) of some types in Go.

Example:
```go
package main

func main() {
    // Use `nil` as zero values
    _ = interface{}(nil)
    _ = (*struct{})(nil)        // () around *struct{} is necessary, otherwise Go will not be able to know where `*` points to.
    _ = string[](nil)
    _ = map[string]int(nil)
    _ = chan string(nil)
    _ =(func())(nil)            // () around func() is necessary

    // This lines are equivalent to the above lines
    var _ interface{} = nil
    var _ *struct{} = nil
    var _ string[] = nil
    var _ map[string]int = nil
    var _ chan string = nil
    var _ func() = nil

    // This lines are equivalent to the above lines, as `nil` is default values of these types
    var _ interface{}
    var _ *struct{}
    var _ string[]
    var _ map[string]int
    var _ chan string
    var _ func()
}
```



## Using `nil` in Comparison


### Two `nil` Values of Two Different Types Are Not Comparable

Example:
```go
package main

func main() {
    var _ = (*bool)(nil) == (*string)(nil) // compiler error: mismatched types.
    var _ = chan int(nil) == chan bool(nil) // compiler error: mismatched types.
}
```

These code will fail to compile as they are trying to compare `nil` values of two different types.


### Two `nil` Values of The Same Type May Not Be Comparable

Example:
```go
package main

func main() {
    var _ = ([]string)(nil) == ([]string)(nil)                  // compiler error: invalid operation.
    var sb = (map[string]bool)(nil) == (map[string]bool)(nil)   // compiler error: invalid operation.
    var _ = (func())(nil) == (func())(nil)                      // compiler error: invalid operation.
}
```

Take `var sb = (map[string]bool)(nil) == (map[string]bool)(nil)` as an example, the reason why two `nil` values of a same type (`map[string]bool`) are not comparable is because Go does not support comparison in slice, map and function types. **You can see that we are comparing two values of a non-comparable type in this case. That is why it fails.**

But the following code works and results are true:

```go
    var _ = ([]string)(nil) == nil              // true
    var sb = (map[string]bool)(nil) == nil      // true
    var _ = (func())(nil) == nil                // true
```

Take `var sb = (map[string]bool)(nil) == nil` as an example, `(map[string]bool)(nil)` declares a `map[string]bool` temporary variable which value is `nil` and `(map[string]bool)(nil) == nil` detects whether the variable's value is `nil` and then assigns the results to `sb`. **You can see that we are comparing the value of a non-comparable type with its zero value (`nil`) in this case. That's why it works.**


### Two `nil` Values of The Same Type Can Be Comparable Only When This Type Supports Comparision

Example:
```go
package main
import "fmt"

func main() {
    fmt.Println( interface{}(nil) == interface{}(nil) ) // true
    fmt.Println( (*int)(nil) == (*int)(nil) )           // true
    fmt.Println( chan string(nil) == chan string(nil) ) // true
}
```

### Be Careful in `nil` Comparision When Interface Values Are Involved

The following code will not cause any compiler failure but the result is `false` other than `true`.

```go
package main
import "fmt"

func main() {
    fmt.Println( interface{}(nil) == (*int)(nil) )      // false
}
```

Explanation:

- **An interface value consists of a dynamic type and a dynamic value.** `interface{}(nil)` declares an interface value with `{type: nil, value: nil}`.
- The non-interface value is converted to the type of the interface value before making the comparison with an interface value. In this example, `(*int)(nil)` is converted to an interface value with `{type: *int, value: nil}`.
- Two `nil` interface values are equivalent only when they carry the same type. In this case, the converted interface value `{type: *int, value: nil}` has a concrete dynamic type but the other interface value has not. That is why the comparison result is `false`.

A more interesting example:
```go
package main

import "io"
import "bytes"
import "fmt"

func main() {
    var w io.Writer
    fmt.Println(w == nil)   // True

    var b *bytes.Buffer
    w = b
    fmt.Println(w == nil)   // false

    write(w)                // panic: runtime error: invalid memory address or nil pointer dereference
}

// If out is non-nil, output will be written to it.
//
func write(out io.Writer) {
    // ...do something...
    if out != nil {                     // This guard is not secure enough
        out.Write([]byte("done!\n"))
    }
}
```

Explanation:
- **An interface value equals to `nil` only when its type and value are both `nil`.** In example, `w` is an `io.Writer` interface value with `{type: *bytes.Buffer, value: nil}` after the `w = b` assignment. Therefore, `w == nil` is `false` as it carries `*bytes.Buffer` other than `nil` as its concrete dynamic type.

## Summary
- `nil` is and a pre-declared identifier which can be used to represent the zero values of some types in Go.
- Be careful when using `nil` in comparison, especially when interface values are involved. You need to understand what you are comparing: types, or values, or both.
- `(a thing)(nil)` may not equal to `nil`, depends on what that thing is (a pointer or an interface). This means Go is a strong-type language and it also applies to `nil` even though `nil` itself does not have default type (**sarcasm**).
