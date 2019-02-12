package v1

import (
	"fmt"
)

// Create - the implementation of the `Create` method. It uses builtin errors to do the error handling
func (m *manager) Create(firstName, lastName, password, email string) (string, error) {
	var ID string

	if `the password containers some characters that the system cann't recognize` {
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

// Create - the implementation of the `Create` method. It uses the first solution to do the error handling.
func (m *manager) Create(firstName, lastName, password, email string) (string, error) {
	var ID string

	if `the password containers some characters that the system cann't recognize` {
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


// Create - the implementation of the `Create` method. It uses the second solution to do the error handling.
func (m *manager) Create(firstName, lastName, password, email string) (string, error) {
	var ID string

	if `the password containers some characters that the system cann't recognize` {
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
