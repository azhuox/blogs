# Error Handling In Golang

What is included in this blog:

- A discussion about how to do error handling in Golang

## prerequisites
I recommend you read [this doc](https://golang.org/doc/faq#nil_error) to understand why it is recommened for functions that return errors always to use `error` interface (defined in `$GOROOT/src/builtin`) other than concrete error types in their signature.

## Use Case

Suppose you have a micro-service called `users` which is used to manage users in a system and you are write an API for creating users in the system.

Here is the API specification

```go
Create a user: POST https://user.micro-service.com/users/v1/ json:{FirstName: string, LastName: string, Email: string, Phone: string}
```

### Pseudocode

Here is the pseudocode code of the `Create` method in `userManager`:

```go
// Create - the implementation of the `Create` method. It uses builtin errors to do the error handling
func (m *manager) Create(firstName, lastName, password, email, phone string) (string, error) {
	var ID string

	if `the password containers some characters that the system cann't recognize` {
		return ID, fmt.Errorf("The password contains some invalid characters.") 
	}

	if `a user with the given email already exists` {
		return ID, fmt.Errorf("The email %s has been used by another user.", email) 
	}
	
	ID, err := mysqlClient.CreateUser(firstName, lastName, password, email, phone)
	if err != nil {
		return ID, fmt.Errorf("Error creating user {Name: %s %s, Email: %s}, err: %s", firstName, lastName, email, err.Error())
	}

	return ID, nil
}
```

Here is the pseudocode code of the API handler:

```go
// CreateUserAPIHandler is the API handler for creating a site
func CreateUserAPIHandler(w http.ResponseWriter, r *http.Request) {
    var err error
    var user := &struct{
        Firstname string   `json:"firstname"`
        Lastname  string   `json:"lastname"`
        Pasword string `json:"phone"`
        Email   string `json:"email"`
    }{}

    // Parse args
    params := mux.Vars(r)
    if err = json.NewDecoder(r.Body).Decode(&user); err != nil {
        http.Error(w, fmt.Sprintf("Error decoding request params, err: %s", err.Error()), http.StatusBadRequest)
        return
    }

    // Create a user manager
    userManager, err := user_v1.NewManager(...)
    if err != nil {
        log.Errorf("[user_create_v1] error creating user manager, err: %s", err.Error())
        http.Error(w, fmt.Sprintf("Internal server error, please retry later"), http.StatusBadRequest)
        return
    }

    // Use the user manager to create a user with given parameters
    if err != userManager.Create(user.FirstName, user.LastName, user.Email, user.Phone) {
        log.Errorf("[user_create_v1] error creating the user %#v, err: %s", user, err.Error())
        http.Error(w, fmt.Sprintf("Internal server error, please retry later"), http.StatusInternalServerError)
        return
    }

    // Return ID
    json.NewEncoder(w).Encode(&struct{ID string `json:"ID"`}{ID: ID})
    w.WriteHeader(http.StatusOK)
    return
}
```
## Error Handling

Everything works fine in these peices of code. However, one thing you may not be happy is that the API handler always returns `http.StatusInternalServerError` as status code for any error returned by `userManager.Create`. As some of the errors are not internal server errors. Moreover, you want to return different status code based on different errors. The key point is to let `userManager.Create` return specific error types and let the API handler set correct status code based on these errors.

### Solution 1: Utilize Golang Structs to Define Specific Error Types

The first solution is to utilize a struct to represent a specific type of error. Here is an example:
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
// Create - the implementation of the `Create` method. It uses the first solution to do the error handling.
func (m *manager) Create(firstName, lastName, password, email string) (string, error) {
	var ID string

	if `the password containers some characters that the system cann't recognize` {
		return ID, newBadRequestErr("The password contains some invalid characters.") 
	}

	if `a user with the given email already exists` {
		return ID, newEmailHasBeenUsedErr("The email %s has been used by another user.", email) 
	}
	
	ID, err := mysqlClient.CreateUser(firstName, lastName, password, email)
	if err != nil {
		return ID, newInternelServerErr("Error creating user {Name: %s %s, Email: %s}, err: %s", firstName, lastName, email, err.Error())
	}

	return ID, nil
}
```

Error handling in the API handler:

```
// CreateUserAPIHandler is the API handler for creating a site. It uses the first solution to do the error handling.
func CreateUserAPIHandler(w http.ResponseWriter, r *http.Request) {
    var err error
    ...

    // Use the user manager to create a user with given parameters
    ID, err := userManager.Create(user.FirstName, user.LastName, user.Pasword, user.Email)
    if err != nil {
		log.Errorf("[user_create_v1] error creating the user %#v, err: %s", user, err.Error())
		if _, ok := err.(*user_v1.BadRequestErr); ok {
			http.Error(w, fmt.Sprintf("Bad request: %s", err.Error()), http.StatusBadRequest)
		} else if _, ok := err.(*user_v1.EmailHasBeenUsedErr); ok {
			http.Error(w, fmt.Sprintf("Bad request: %s", err.Error()), http.StatusConflict)
		} else if _, ok := err.(*user_v1.InternelServerErr); ok {
			http.Error(w, fmt.Sprintf("Internal server error, please retry later."), http.StatusInternalServerError)
		} else {
            http.Error(w, fmt.Sprintf("Unknown error, please retry later."), http.StatusInternalServerError)
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

- Return an `error` interface other than specific error types in the `userManager.Create` function's signature. You can read [this doc](https://golang.org/doc/faq#nil_error) for the reason. 
- Expose those error types so that the caller can do the erro handling by eror convertion. 
- Do not expose `new` methods of those error types make them read-only. Moreover, return an `error` interface as well in those `new` methods' signature. This will convert an error type's pointer (say `*user_v1.BadRequestErr`) to the `erro` interface.
- Do the error handling by converting an `err` interface to a specific erro type's pointer. The reason why this works is because an `err` returned by the `userManager.Create` method is essentially an `error` interface with avalue and a specific erro type (say `*user_v1.BadRequestErr`). Therefore, `if _, ok := err.(*user_v1.BadRequestErr); ok {...}` totally works. 

### Advantage
- It does not break the princeple of returning an `error` interface in functions' signature.
- It provides a way for you to handle different errors. 
- Each error type can be customized. This allows you to provide you more details about an error. The `SyntaxError` error type in Golang [json package](https://golang.org/pkg/encoding/json/) is a perfect example. It has a memeber called `Offset` which is used to where the error occured after reading bytes. Here is the `SyntaxError` error type definition: 
```go
type SyntaxError struct {
    msg    string // description of error
    Offset int64  // error occurred after reading Offset bytes
}

func (e *SyntaxError) Error() string { return e.msg }
```
How to use the `Offset` member:
```go
if err := dec.Decode(&val); err != nil {
    if serr, ok := err.(*json.SyntaxError); ok {
        line, col := findLine(f, serr.Offset)
        return fmt.Errorf("%s:%d:%d: %v", f.Name(), line, col, err)
    }
    return err
}
```

### Disadventage
- The conversion from the `error` interface to specific error types is ugly, especially when you need to this several times. You will have a long `if else` statement once you need to do such conversion several times. Moreover, `switch case` statement will not be to used in this case, as the `err.(type)`  from the `userManager.Create` method is always an `error` interface ohter than those specific error types. 
- Defining those error types and `new` methods is somehow overwhelmed. You can see that you need to define a struct and a `new` method for every error type. Plus, from the example you can see that, in some cases,  we don't use an error type's methods or members, instead we only care about what error type it is. It is overwhelmed to define an error type with a struct just for achieving this goal.
- I personally don't like the idea of converting an `error` interface back to a specific error type. First, it somehow forces users to figure out whehter an error type or the error type's pointer is used. For example, `if _, ok := err.(*user_v1.BadRequestErr); ok {...}` will not be `OK` if the `userManager.Create` method returns an `BadRequestErr` instead of `*BadRequestErr`. This is because Golang is a strong type language so `BadRequestErr != *BadRequestErr`. Second, in my opinion, an interface is not supposed to converted back to a specific type. This is becasue a Golang interface is designed to focus on some behaviours (methods defined in the interface) and ignore the implementation details. Converting an `error` interface back to a specific error type means you want to expose some implemenation details, thus violating the priinceple I just mentioned.      


### Solution 2: Define Error Type as A Property

Instead of using Golang strcuts to define error types, the second solution adds and `Type()` method to return error types in a customize `Error` interface extended from the `error` interface. Here is the definition of the `Error` interface and its implementation:

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
```

There is the definiton of specific error types. You can see that strcuts in the first solution are replaced with constants in this solution.

```go
// ErrType - error type
type ErrType string

// Bad request errors
const (
	// ErrTypeBadRequest - bad request
	ErrTypeBadRequest          ErrType = "bad_request"
	// ErrTypeEmailHasBeenUsed - email has been used
	ErrTypeEmailHasBeenUsed   ErrType = "email_has_been_used"
	// ErrTypeInternalServerErr - internal server error
	ErrTypeInternalServerErr       ErrType = "internal_server_error"
	// ErrTypeUnknown - Unknown error
	ErrTypeUnknown ErrType = "unknown"
)
```

How to use these error types:

```go
// Create - the implementation of the `Create` method. It uses the second solution to do the error handling.
func (m *manager) Create(firstName, lastName, password, email string) (string, error) {
	var ID string

	if `the password containers some characters that the system cann't recognize` {
		return ID, newError(ErrTypeBadRequest, "The password contains some invalid characters.") 
	}

	if `a user with the given email already exists` {
		return ID, newError(ErrTypeEmailHasBeenUsed, "The email %s has been used by another user.", email) 
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
func CreateUserAPIHandler(w http.ResponseWriter, r *http.Request) {
    ...

    // Use the user manager to create a user with given parameters
	ID, err := userManager.Create(user.FirstName, user.LastName, user.Pasword, user.Email); 
	if err != nil {
		log.Errorf("[user_create_v1] error creating the user %#v, err: %s", user, err.Error())

		if userManErr, ok := err.(*user_v1.Error); ok {
			// Upgrade an `error` interface to a `user_v1.Error` interface so that we can use the `Type()` method to get the error type
			switch userManErr.Type() {
			case user_v1.ErrTypeBadRequest:
				http.Error(w, fmt.Sprintf("Bad request: %s", userManErr.Error()), http.StatusBadRequest)
			case user_v1.ErrTypeEmailHasBeenUsed:
				http.Error(w, fmt.Sprintf("Bad request: %s", err.Error()), http.StatusConflict)
			case user_v1.ErrTypeInternalServerErr:
				http.Error(w, fmt.Sprintf("Internal server error, please retry later."), http.StatusInternalServerError)
			default:
				http.Error(w, fmt.Sprintf("Unknown error, please retry later."), http.StatusInternalServerError) 
			}
		} else {
			// This should never happen
			http.Error(w, fmt.Sprintf("Unknown error, please retry later."), http.StatusInternalServerError)
		}
	}
	// Return ID
    json.NewEncoder(w).Encode(&struct{ID string `json:"ID"`}{ID: ID})
    w.WriteHeader(http.StatusOK)
    return
}
```

### Key Points
- **The idea behind the second solution is extend the `error` interface with a `Type()` method which is used to return error types**
- The cutomize interface `user_v1.Error` other than pointers of specific error types is returned underneath. 
- The implementation of the `user_v1.Error` interface is hided, only the interface gets exposed. 
- The Golang structs in the first solution are replaced with the constants to represents error types. The callers of the `userManager.Create()` method  utilizes these constants to do the error handling.
- an `error` interface is upgraded to an `user_v1.Error` interface for parsing the errors returned by the `userManager.Create()` method.

### Advantage
- Return an `error` interface in the `userManager.Create()` method's signature allows you to return either a regular error or a `user_v1.Error`, although you should always return `user_v1.Error` for any methods in the `user_v1` package. 
- 
- It hides 