package v1

import (
	"fmt"
)

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

// EmailHasBeenUsed represents resource conflict errors
type ConflictErr struct {
	*baseErr
}

// newConflictErr creates an instance of ConflictErr
func newConflictErr(format string, a ...interface{}) error {
	return &ConflictErr {
		baseErr: newBaseErr(format, a...),
	}
}

// InternelServerErr represents internal server errors
type InternelServerErr struct {
	*baseErr
}

// newInternelServerErr creates an instance of InternelServerErr
func newInternelServerErr(format string, a ...interface{}) error {
	return &InternelServerErr {
		baseErr: newBaseErr(format, a...),
	}
}

