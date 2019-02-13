# Error Handling In Golang

What is included in this blog:

- A discussion about how to do error handling in Golang.

## prerequisites
Before reading this blog, I recommend you reading [this doc](https://golang.org/doc/faq#nil_error) to understand why it is recommended for functions that return errors always to use the `error` interface (defined in `$GOROOT/src/builtin`) other than concrete error types in their signature.

## Use Case

Suppose you have a micro-service called `users-usvc` which is used to manage users in a system and you are adding an API to the micro-service for creating users in the system.

Here is the API specification

```go
POST https://user.micro-service.com/users/v1/
Request Body:{FirstName: string, LastName: string, Password: string, Email: string}
```

### Pseudocode

Here is the pseudocode code of the `Create` method in the `userManager` object (which is used to manage users in the database):

```go
// Create - the implementation of the `Create` method. It uses builtin errors to do the error handling
func (m *manager) Create(firstName, lastName, password, email string) (string, error) {
  var ID string

  if `the password containers some characters that the system can't recognize` {
     return ID, fmt.Errorf("The password contains some invalid characters.")
  }

  if `a user with the given email already exists` {
     return ID, fmt.Errorf("The email %s has been used by another user.", email)
  }

  ID, err := mysqlClient.CreateUser(firstName, lastName, password, email)
  if err != nil {
     return ID, fmt.Errorf("Error creating user {Name: %s %s, Email: %s}, err: %s", firstName, lastName, email, err.Error())
  }

  return ID, nil
}
```

Here is the pseudocode code of the API handler:

```go

import (
    ...
    userV1 "github.com/aaronzhuo1990/blogs/golang/error_handling/users-usvc/internal/user/v1"
)

// CreateUserAPIHandler is the API handler for creating a site. It uses builtin errors to do the error handling
func CreateUserAPIHandler(w http.ResponseWriter, r *http.Request) {
   var err error
   user := &struct{
       FirstName string   `json:"firstname"`
       LastName  string   `json:"lastname"`
       Password string `json:"phone"`
       Email   string `json:"email"`
   }{}

   // Parse args
   if err = json.NewDecoder(r.Body).Decode(&user); err != nil {
       http.Error(w, fmt.Sprintf("Error decoding request params, err: %s", err.Error()), http.StatusBadRequest)
       return
   }

   // Create a user manager
   userManager, err := userV1.NewManager(...)
   if err != nil {
       log.Printf("[user_create_v1] error creating user manager, err: %s", err.Error())
       http.Error(w, "Internal server error, please retry later", http.StatusBadRequest)
       return
   }

   // Use the user manager to create a user with given parameters
   ID, err := userManager.Create(user.FirstName, user.LastName, user.Password, user.Email)
  if err != nil {
       log.Printf("[user_create_v1] error creating the user %#v, err: %s", user, err.Error())
       http.Error(w, "Internal server error, please retry later", http.StatusInternalServerError)
       return
     }
  // Return ID
   json.NewEncoder(w).Encode(&struct{ID string `json:"ID"`}{ID: ID})
   w.WriteHeader(http.StatusOK)
   return
}
```

## Error Handling

Everything works fine in these code. However, one thing you may not be happy about is that the API handler always returns `http.StatusInternalServerError` as status code for any error returned by the `userManager.Create()` method. You may want to return different status codes based on different error types. The key point to solve this problem is to let the `userManager.Create()` method return specific error types and let the API handler set the correct status code based on these errors.

### Solution 1: Utilize Golang Structs to Define Specific Error Types

The first solution is utilizing Golang structs to define error types. Here is an example:

```go
// baseErr - base class
type baseErr struct {
  msg string
}

// Error implements the `Error` method defined in error interface
func (e *baseErr) Error() string {
  if e != nil {
     return e.msg
  }
  return ""
}

// newBaseErr creates an instance of internal error
func newBaseErr(format string, a ...interface{}) *baseErr {
  return &baseErr {
     msg: fmt.Sprintf(format, a...),
  }
}

// BadRequestErr represents bad request errors
type BadRequestErr struct {
  *baseErr
}

// newBadRequestErr creates an instance of BadRequestErr
func newBadRequestErr(format string, a ...interface{}) error {
  return &BadRequestErr {
     baseErr: newBaseErr(format, a...),
  }
}
```

How to use these error types:

```go
// Create - the implementation of the `Create` method. It uses the first solution to do error handling.
func (m *manager) Create(firstName, lastName, password, email string) (string, error) {
  var ID string

  if `the password containers some characters that the system can't recognize` {
     return ID, newBadRequestErr("The password contains some invalid characters.")
  }

  if `a user with the given email already exists` {
     return ID, newConflictErr("The email %s has been used by another user.", email)
  }

  ID, err := mysqlClient.CreateUser(firstName, lastName, password, email)
  if err != nil {
     return ID, newInternelServerErr("Error creating user {Name: %s %s, Email: %s}, err: %s", firstName, lastName, email, err.Error())
  }

  return ID, nil
}
```

Error handling in the API handler:

```go
// CreateUserAPIHandler is the API handler for creating a site. It uses the first solution to do the error handling.
func CreateUserAPIHandler(w http.ResponseWriter, r *http.Request) {
  var err error

  ... // A bunch of operations are omitted

  // Use the user manager to create a user with given parameters
  ID, err := userManager.Create(user.FirstName, user.LastName, user.Password, user.Email)
  if err != nil {
     log.Printf("[user_create_v1] error creating the user %#v, err: %s", user, err.Error())
     if _, ok := err.(*userV1.BadRequestErr); ok {
        http.Error(w, fmt.Sprintf("Bad request: %s", err.Error()), http.StatusBadRequest)
     } else if _, ok := err.(*userV1.ConflictErr); ok {
        http.Error(w, fmt.Sprintf("Bad request: %s", err.Error()), http.StatusConflict)
     } else if _, ok := err.(*userV1.InternelServerErr); ok {
        http.Error(w, "Internal server error, please retry later.", http.StatusInternalServerError)
     } else {
        // This should never happen
          http.Error(w, "Unknown error, please retry later.", http.StatusInternalServerError)
      }
      return
  }
  // Return ID
  json.NewEncoder(w).Encode(&struct{ID string `json:"ID"`}{ID: ID})
  w.WriteHeader(http.StatusOK)
  return
}
```


### Key Points

- Return an `error` interface other than specific error types in the signature of the `userManager.Create()` method. You can read [this doc](https://golang.org/doc/faq#nil_error) for the reason.
- Expose customized error types so that callers can do the error handling by converting an `error` interface to a specific error type.
- Do not expose the `new` methods of those error types in order to make them read-only. Moreover, return an `error` interface in the signature of these`new` methods' as well, as this converts an error type's pointer (say `*userV1.BadRequestErr`) to an `error` interface.
- Do the error handling by converting an `error` interface to a specific error type's pointer. The reason why this works is that the `err` returned by the `userManager.Create()` method is essentially an `error` interface with a value and a type which essentially is an error type’s pointer (say `*userV1.BadRequestErr`). Therefore, `if _, ok := err.(*userV1.BadRequestErr); ok {...}` totally works as it just converts the `err` back to its type.

### Pros
- It follows  the principle of returning an `error` interface in a function's signature.
- It provides a way for you to handle specific errors.
- Each error type can be customized. This allows you to add more details to an error type. The `SyntaxError` error type in Golang [json package](https://golang.org/pkg/encoding/json/) is a perfect example. It has a member called `Offset` which is used to indicate where the error occurred after reading bytes. Here is the `SyntaxError` error type definition:
```go
type SyntaxError struct {
   msg    string // description of error
   Offset int64  // error occurred after reading Offset bytes
}

func (e *SyntaxError) Error() string { return e.msg }
```

Here is an example of using the `Offset` member:

```go
if err := dec.Decode(&val); err != nil {
   if serr, ok := err.(*json.SyntaxError); ok {
       line, col := findLine(f, serr.Offset)
       return fmt.Errorf("%s:%d:%d: %v", f.Name(), line, col, err)
   }
   return err
}
```

### Cons
- The conversion from an `error` interface to a specific error type is ugly, especially when you need to this several times. You will have a long `if else` statement if you have to handle a bunch of error types. The `switch case` statement will not be able to be used in this case, as the `err.(type)` from the `userManager.Create()` method is always an `error` interface other than those specific error types.
- Defining those error types and `new` methods are somehow overwhelmed. You will need to crate a struct and a `new` method for every error type. Plus, from the example, you can see that, in some cases,  we don't use an error type's methods or members, instead we only care about what the error type is. It is overwhelmed to use a struct to define an error type just for achieving this goal.
- I personally don't like the idea of converting an `error` interface back to a specific error type. First, it somehow forces callers to figure out whether an error type or the error type's pointer is actually returned. For example, `if _, ok := err.(*userV1.BadRequestErr); ok {...}` will not work if the `userManager.Create()` method returns a `BadRequestErr` instead of `*BadRequestErr`. This is because Golang is a strong type language, so `BadRequestErr` does not eqaul to *BadRequestErr`. Second, in my opinion, an interface is not supposed to be converted back to a specific type. This is because a Golang interface is designed for you to focus on some behaviors (which are methods defined in the interface) and ignore the implementation details. Converting an `error` interface back to a specific error type means you want to expose some implementation details, thus violating the principle I just mentioned.


### Solution 2: Define Error Type as A Property

Instead of using Golang structs to define error types, the second solution adds  a `Type()` method to return error types in a customize `userV1.Error` interface extended from the `error` interface. Here is the definition of the `userV1. Error` interface and its implementation:

```go
// Error interface defines the errors used in this package
type Error interface {
  error
  Type() ErrType
}

// errorImpl - implementation of Error interface
type errImpl struct {
  msg     string
  errType ErrType
}

// Error returns error message
func (e *errImpl) Error() string {
  if e != nil {
     return e.msg
  }
  return ""
}

// Type returns error type
func (e *errImpl) Type() ErrType {
  if e != nil {
     return e.errType
  }
  return ErrTypeUnknown
}

// newError returns an error with given error type
func newError(errType ErrType, format string, a ...interface{}) Error {
  return &errImpl{
     msg: fmt.Sprintf(format, a...),
     errType: errType,
  }
}

// ConvertError - try converting an `error` interface to an `Error` interface
func ConvertError(err error) (Error, bool) {
  if e, ok := err.(Error); ok {
     return e, ok
  }

  return nil, false
}

```

Here is the definition of customized error types. You can see that the Golang structs defined in the first solution are replaced with constants in this solution.

```go
// Bad request errors
const (
  // ErrTypeBadRequest - bad request
  ErrTypeBadRequest          ErrType = "bad_request"
  // ErrTypeConflict - resource conflicts
  ErrTypeConflict ErrType = "conflict"
  // ErrTypeInternalServerErr - internal server error
  ErrTypeInternalServerErr       ErrType = "internal_server_error"
  // ErrTypeUnknown - Unknown error
  ErrTypeUnknown ErrType = "unknown"
)
```

How to use these error types:

```go
// Create - the implementation of the `Create` method. It uses the second solution to do error handling.
func (m *manager) Create(firstName, lastName, password, email string) (string, error) {
  var ID string

  if `the password containers some characters that the system can't recognize` {
     return ID, newError(ErrTypeBadRequest, "The password contains some invalid characters.")
  }

  if `a user with the given email already exists` {
     return ID, newError(ErrTypeConflict, "The email %s has been used by another user.", email)
  }

  ID, err := mysqlClient.CreateUser(firstName, lastName, password, email)
  if err != nil {
     return ID, newError(ErrTypeInternalServerErr, "Error creating user {Name: %s %s, Email: %s}, err: %s", firstName, lastName, email, err.Error())
  }

  return ID, nil
}
```

Error handling in the `create a user` API handler

```go
// CreateUserAPIHandler is the API handler for creating a site. It uses the second solution to do error handling.
func CreateUserAPIHandler(w http.ResponseWriter, r *http.Request) {
  var err error

  ... // A bunch of operations are omitted

   // Use the user manager to create a user with given parameters
  ID, err := userManager.Create(user.FirstName, user.LastName, user.Password, user.Email);
  if err != nil {
     log.Printf("[user_create_v1] error creating the user %#v, err: %s", user, err.Error())

     if uErr, ok := userV1.ConvertError(err); ok {
        // Upgrade an `error` interface to a `userV1.Error` interface so that we can use the `Type()` method to get the error type
        switch uErr.Type() {
        case userV1.ErrTypeBadRequest:
           http.Error(w, fmt.Sprintf("Bad request: %s", uErr.Error()), http.StatusBadRequest)
        case userV1.ErrTypeConflict:
           http.Error(w, fmt.Sprintf("Bad request: %s", err.Error()), http.StatusConflict)
        case userV1.ErrTypeInternalServerErr:
           http.Error(w, "Internal server error, please retry later.", http.StatusInternalServerError)
        default:
           http.Error(w, "Unknown error, please retry later.", http.StatusInternalServerError)
        }
     } else {
        // This should never happen
        http.Error(w, "Unknown error, please retry later.", http.StatusInternalServerError)
     }
  }
  // Return ID
   json.NewEncoder(w).Encode(&struct{ID string `json:"ID"`}{ID: ID})
   w.WriteHeader(http.StatusOK)
   return
}
```

### Key Points
- **The idea behind this solution is  extending the `error` interface to a customized `userV1.Error` interface with a `Type()` method which is used to return error types**
- the `userManager.Create()` method uses an `userV1.Error` interface other than an error type's pointer to record errors. This ensures no implementation details of the `userV1.Error` interface gets exposed.
- The signature of `userManager.Create()` method still returns an `error` interface other than the `userV1.Error` interface. This gives you the freedom to keep using the same `error` interface instance `err` that is created at the beginning of the API handler and allows you to do the conversion whenever you need. It is like we provide you with a great feature, but we do not force you to use this feature.
- The Golang structs in the first solution are replaced with the constants to represents error types. Then callers of the `userManager.Create()` method can utilize these constants to do the error handling.
- In the `userV1.ConvertError()` method, an `error` interface is upgraded to an `userV1.Error` interface for parsing an error returned by the `userManager.Create()` method.

### Pros
- Return an `error` interface in the signature of the `userManager.Create()` method allows you to return either a regular `error` instance or a `userV1.Error` instance. (Although you should always return `userV1.Error` for any methods in the `userV1` package.)
- It hides the details of how a private Golang struct (`userV1.errorImpl`) is implemented to fit the `userV1.Error` interface and only exposes what it needs to be exposed.
- It is easier and it has less work to define error types with constants other than structs.

### Cons
- Customizing error types becomes impossible in this solution. This is because all the error types are constructed from the same Golang struct `userV1.errorImpl` and they all follow the constraint of the `userV1.Error` interface.

## Summary
- It is recommended for functions that return errors always to use the `error` interface other than concrete error types in their signature.
- Use structs to define error types and expose them if you need to customize some error types for a package.
- Define a customized interface to extend the `error` interface if all of the error types that you want to define have the same properties.

BTW, you can check the complete example from [this repo](https://github.com/aaronzhuo1990/blogs/tree/master/golang/error_handling/users-usvc).

That's it, thanks for reading this blog.

Reference
- [Why is my nil error value not equal to nil?](https://golang.org/doc/faq#nil_error)
- [Json Package in Golang](https://golang.org/pkg/encoding/json/)
- [Error Handling and Go](https://blog.golang.org/error-handling-and-go)
- [The Complete Golang Error Handling Example in This Blog](https://github.com/aaronzhuo1990/blogs/tree/master/golang/error_handling/users-usvc)

